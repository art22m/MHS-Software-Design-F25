cmake_minimum_required(VERSION 3.8)

project(yagi)

set(CMAKE_CXX_STANDARD 23)

set(CMAKE_EXPORT_COMPILE_COMMANDS ON)

set(CMAKE_CXX_FLAGS "${CMAKE_CXX_FLAGS} -Wall -Werror -Wextra-semi -O1 -g -fsanitize=address -fno-omit-frame-pointer")

add_subdirectory(msglib)

add_subdirectory(server)

add_subdirectory(client)
