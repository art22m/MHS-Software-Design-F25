add_library(msglib STATIC msglib.cpp)

target_include_directories(msglib PUBLIC ${CMAKE_CURRENT_SOURCE_DIR})
