#include <msglib.hpp>

#include <atomic>
#include <cstring>
#include <iostream>
#include <map>
#include <mutex>
#include <set>
#include <shared_mutex>
#include <thread>
#include <unistd.h>
#include <vector>

#include <sys/event.h>
#include <sys/time.h>
#include <sys/types.h>

using namespace msglib;

constexpr size_t MAX_PACKET_SIZE = 4096;

class RoomManager {
private:
  std::shared_mutex mutex_;
  std::map<std::string, std::set<int>> rooms_;
  std::map<int, std::string> clientRoom_;

public:
  void Join(const std::string &room, int fd) {
    {
      std::unique_lock lock(mutex_);
      rooms_[room].insert(fd);
      clientRoom_[fd] = room;
    }
    std::cout << "Client " << fd << " joined room: " << room << std::endl;
  }

  void Leave(int fd) {
    std::optional<std::string> leftRoom = {};
    {
      std::unique_lock lock(mutex_);
      if (clientRoom_.count(fd) > 0) {
        leftRoom = clientRoom_[fd];
        rooms_[*leftRoom].erase(fd);

        if (rooms_[*leftRoom].empty()) {
          rooms_.erase(*leftRoom);
        }
        clientRoom_.erase(fd);
      }
    }
    if (leftRoom.has_value()) {
      std::cout << "Client " << fd << " left room: " << *leftRoom << std::endl;
    }
  }

  void Broadcast(int senderFd, const std::string &msg) {
    std::shared_lock lock(mutex_);
    if (!clientRoom_.count(senderFd)) {
      return;
    }

    const std::string &roomName = clientRoom_.at(senderFd);
    if (!rooms_.count(roomName)) {
      return;
    }

    const auto &participants = rooms_.at(roomName);

    for (int targetFd : participants) {
      if (targetFd == senderFd) {
        continue;
      }

      auto formattedMsg = std::format("{}: {}", targetFd, msg);
      MsgHeader header{static_cast<uint32_t>(formattedMsg.size()),
                       MsgType::TEXT};

      send(targetFd, &header, sizeof(header), 0);
      send(targetFd, formattedMsg.data(), formattedMsg.size(), 0);
    }
  }
};

class Worker {
public:
  Worker(RoomManager &manager) : TheRoomManager(manager) {
    kqueueFd_ = kqueue();
    if (kqueueFd_ == -1)
      throw std::runtime_error("kqueue creation failed");
  }

  ~Worker() {
    running_ = false;
    if (thread_.joinable())
      thread_.join();
    close(kqueueFd_);
  }

  void Start() {
    thread_ = std::thread([this]() { Run(); });
  }

  void AddClient(int fd) {
    std::lock_guard<std::mutex> lock(clientsMutex_);
    clients_[fd] = std::make_unique<Socket>(fd);

    struct kevent ev;
    EV_SET(&ev, fd, EVFILT_READ, EV_ADD | EV_ENABLE, 0, 0, nullptr);

    if (kevent(kqueueFd_, &ev, 1, nullptr, 0, nullptr) == -1) {
      std::cerr << "kevent ADD failed for fd " << fd << std::endl;
    }
  }

private:
  void Run() {
    std::vector<struct kevent> events(64);
    struct timespec timeout;

    while (running_) {
      timeout.tv_sec = 1;
      timeout.tv_nsec = 0;

      int numEvents =
          kevent(kqueueFd_, nullptr, 0, events.data(), events.size(), &timeout);

      if (numEvents == -1) {
        continue;
      }

      for (int i = 0; i < numEvents; ++i) {
        int clientFd = (int)events[i].ident;

        if (events[i].flags & EV_EOF) {
          Disconnect(clientFd);
        } else if (events[i].filter == EVFILT_READ) {
          HandleClient(clientFd);
        }
      }
    }
  }

  void HandleClient(int fd) {
    MsgHeader header;
    if (!clients_.at(fd)->RecvAll(&header, sizeof(header))) {
      Disconnect(fd);
      return;
    }

    if (header.len > MAX_PACKET_SIZE) {
      std::cerr << "Error: Received packet too large from fd " << fd
                << std::endl;
      Disconnect(fd);
      return;
    }

    std::vector<char> buffer(header.len);
    if (header.len > 0) {
      if (!clients_.at(fd)->RecvAll(buffer.data(), header.len)) {
        Disconnect(fd);
        return;
      }
    }

    switch (header.msgType) {
    case MsgType::CONNECT:
      if (header.len > 0) {
        std::string roomName(buffer.begin(), buffer.end());
        TheRoomManager.Join(roomName, fd);
      }
      break;
    case MsgType::TEXT:
      if (header.len > 0) {
        std::string msg(buffer.begin(), buffer.end());
        TheRoomManager.Broadcast(fd, msg);
      }
      break;
    case MsgType::DISCONNECT:
      Disconnect(fd);
      break;
    default:
      std::cerr << "Warning: Unknown message type received from fd " << fd
                << std::endl;
    }
  }

  void Disconnect(int fd) {
    TheRoomManager.Leave(fd);

    struct kevent ev;
    EV_SET(&ev, fd, EVFILT_READ, EV_DELETE, 0, 0, nullptr);
    kevent(kqueueFd_, &ev, 1, nullptr, 0, nullptr);

    std::lock_guard<std::mutex> lock(clientsMutex_);
    clients_.erase(fd);
  }

private:
  RoomManager &TheRoomManager;
  int kqueueFd_;
  std::atomic<bool> running_{true};
  std::thread thread_;
  std::mutex clientsMutex_;
  std::map<int, std::unique_ptr<Socket>> clients_;
};

int main() {
  RoomManager roomManager;

  const size_t kWorkerCount = std::thread::hardware_concurrency();
  std::vector<std::unique_ptr<Worker>> workers;

  for (int i = 0; i < kWorkerCount; ++i) {
    workers.push_back(std::make_unique<Worker>(roomManager));
    workers.back()->Start();
  }

  std::cout << "Server started with " << kWorkerCount << " worker threads."
            << std::endl;

  Socket listener;
  try {
    listener.Bind(8080);
    listener.Listen();
  } catch (const std::exception &e) {
    std::cerr << "Listener setup failed: " << e.what() << std::endl;
    return 1;
  }

  size_t roundRobinIdx = 0;
  while (true) {
    try {
      std::unique_ptr<Socket> client = listener.Accept();

      // Assume Socket::Release() exists to prevent double close
      workers[roundRobinIdx]->AddClient(client->Release());

      roundRobinIdx = (roundRobinIdx + 1) % kWorkerCount;

    } catch (const std::exception &e) {
      std::cerr << "Accept error: " << e.what() << std::endl;
    }
  }
  return 0;
}