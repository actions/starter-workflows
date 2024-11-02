project(HelloWorld)
cmake_minimum_required(VERSION 3.0)

add_executable(hello_world source/main.cpp)

# enable testing functionality
enable_testing()
# define tests
add_test(test1 ./hello_world)