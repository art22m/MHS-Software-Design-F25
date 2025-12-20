#include <cstdint>
#include <memory>
#include <stdexcept>
#include <string>

#include <arpa/inet.h>
#include <netinet/in.h>
#include <sys/socket.h>
#include <unistd.h>

namespace msglib {
enum class MsgType : std::uint8_t {
  TEXT = 0,
  CONNECT = 1,
  DISCONNECT = 2,
};

struct MsgHeader {
  std::uint32_t len;
  MsgType msgType;
};

constexpr size_t kMaxPacketSize = 4096;

class Socket {
public:
  Socket() {
    fd_ = socket(AF_INET, SOCK_STREAM, 0);
    if (fd_ == -1)
      throw std::runtime_error("socket creation failed");
    int opt = 1;
    setsockopt(fd_, SOL_SOCKET, SO_REUSEADDR, &opt, sizeof(opt));
  }

  explicit Socket(int fd) : fd_(fd) {}

  ~Socket() {
    if (fd_ != -1) {
      close(fd_);
    }
  }

  Socket(const Socket &) = delete;
  Socket &operator=(const Socket &) = delete;
  Socket(Socket &&) = delete;
  Socket &operator=(Socket &&) = delete;

  void Bind(int port) {
    sockaddr_in addr{};
    addr.sin_family = AF_INET;
    addr.sin_addr.s_addr = INADDR_ANY;
    addr.sin_port = htons(port);
    if (bind(fd_, (struct sockaddr *)&addr, sizeof(addr)) < 0) {
      throw std::runtime_error("bind failed");
    }
  }

  void Listen(int backlog = 10) {
    if (listen(fd_, backlog) < 0) {
      throw std::runtime_error("listen failed");
    }
  }

  std::unique_ptr<Socket> Accept() {
    sockaddr_in addr{};
    socklen_t len = sizeof(addr);
    int client_fd = accept(fd_, (struct sockaddr *)&addr, &len);
    if (client_fd < 0) {
      throw std::runtime_error("accept failed");
    }

    return std::make_unique<Socket>(client_fd);
  }

  void Connect(const std::string &host, int port) {
    sockaddr_in addr{};
    addr.sin_family = AF_INET;
    addr.sin_port = htons(port);
    if (inet_pton(AF_INET, host.c_str(), &addr.sin_addr) <= 0) {
      throw std::runtime_error("invalid address");
    }
    if (connect(fd_, (struct sockaddr *)&addr, sizeof(addr)) < 0) {
      throw std::runtime_error("connect failed");
    }
  }

  bool SendAll(const void *data, size_t len) {
    const char *ptr = static_cast<const char *>(data);
    size_t total = 0;
    while (total < len) {
      ssize_t sent = send(fd_, ptr + total, len - total, 0);
      if (sent <= 0) {
        return false;
      }
      total += sent;
    }

    return true;
  }

  bool RecvAll(void *data, size_t len) {
    char *ptr = static_cast<char *>(data);
    size_t total = 0;
    while (total < len) {
      ssize_t recvd = recv(fd_, ptr + total, len - total, 0);
      if (recvd <= 0) {
        return false;
      }
      total += recvd;
    }
    return true;
  }

  int GetFd() const { return fd_; }

  int Release() {
    int tmp = fd_;
    fd_ = -1;
    return tmp;
  }

private:
  int fd_ = -1;
};

} // namespace msglib