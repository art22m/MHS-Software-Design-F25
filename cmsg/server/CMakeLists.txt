add_executable(msgserver server.cpp)

target_link_libraries(msgserver PRIVATE msglib)
