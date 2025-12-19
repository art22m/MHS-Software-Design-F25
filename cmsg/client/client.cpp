#include <cassert>
#include <msglib.hpp>

#include <cstdlib>
#include <cstring>
#include <iostream>
#include <string>
#include <thread>
#include <vector>

using namespace msglib;

constexpr int kServerPort = 8080;
constexpr char kServerHost[] = "127.0.0.1";

bool SendMessage(Socket &socket, MsgType type, const std::string &payload) {
  assert(payload.size() < kMaxPacketSize);

  MsgHeader header;
  header.msgType = type;
  header.len = static_cast<std::uint32_t>(payload.size());

  if (!socket.SendAll(&header, sizeof(header))) {
    std::cerr << "Connection closed while sending header." << std::endl;
    return false;
  }

  if (payload.size() > 0) {
    if (!socket.SendAll(payload.data(), payload.size())) {
      std::cerr << "Connection closed while sending payload." << std::endl;
      return false;
    }
  }

  return true;
}

void ReceiveLoop(Socket &socket) {
  while (true) {
    MsgHeader header;

    if (!socket.RecvAll(&header, sizeof(header))) {
      std::cout << "\n--- Server disconnected. Press Enter to exit. ---"
                << std::endl;
      break;
    }

    std::vector<char> buffer(header.len);
    if (header.len > 0) {
      if (!socket.RecvAll(buffer.data(), header.len)) {
        std::cout << "\n--- Server disconnected during payload. ---"
                  << std::endl;
        break;
      }
    }

    // 3. Process
    if (header.msgType == MsgType::TEXT) {
      std::string message(buffer.begin(), buffer.end());
      std::cout << "\n[Room Message]: " << message << std::endl;
      std::cout << "You: ";
      std::fflush(stdout); // Ensure prompt shows immediately
    } else if (header.msgType == MsgType::DISCONNECT) {
      std::cout << "\n--- Received DISCONNECT command from server. ---"
                << std::endl;
      break;
    }
  }
  std::exit(0);
}

int main(int argc, char *argv[]) {
  if (argc != 2) {
    std::cerr << "Usage: " << argv[0] << " <RoomName>" << std::endl;
    return 1;
  }
  std::string roomName = argv[1];

  try {
    Socket clientSocket;
    clientSocket.Connect(kServerHost, kServerPort);
    std::cout << "Connected to server. Joining room: " << roomName << std::endl;

    if (!SendMessage(clientSocket, MsgType::CONNECT, roomName)) {
      return 1;
    }

    std::thread receiverThread(ReceiveLoop, std::ref(clientSocket));
    receiverThread.detach();

    std::string input;
    std::cout << "Type messages and press Enter. Type '/exit' to quit."
              << std::endl;
    std::cout << "You: ";

    while (std::getline(std::cin, input)) {
      if (input == "/exit") {
        SendMessage(clientSocket, MsgType::DISCONNECT, "");
        break;
      }

      if (!SendMessage(clientSocket, MsgType::TEXT, input)) {
        break;
      }
      std::cout << "You: ";
    }

  } catch (const std::exception &e) {
    std::cerr << "Client error: " << e.what() << std::endl;
    return 1;
  }

  std::cout << "Exiting client application." << std::endl;
}
