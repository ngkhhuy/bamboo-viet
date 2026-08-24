cmake_minimum_required(VERSION 3.10)
project(fcitx5-bamboo-viet VERSION 1.0.0 LANGUAGES CXX C)

find_package(ECM 1.0.0 REQUIRED NO_MODULE)
set(CMAKE_MODULE_PATH ${ECM_MODULE_PATH} "${CMAKE_CURRENT_SOURCE_DIR}/cmake" ${CMAKE_MODULE_PATH})

find_package(Fcitx5Core REQUIRED)
find_package(Fcitx5Utils REQUIRED)

include(GNUInstallDirs)

# Compile libvicore.so with Go
add_custom_command(
    OUTPUT ${CMAKE_CURRENT_BINARY_DIR}/libvicore.so ${CMAKE_CURRENT_BINARY_DIR}/libvicore.h
    COMMAND go build -buildmode=c-shared -o ${CMAKE_CURRENT_BINARY_DIR}/libvicore.so ${CMAKE_CURRENT_SOURCE_DIR}/../libvicore/
    WORKING_DIRECTORY ${CMAKE_CURRENT_SOURCE_DIR}/..
    COMMENT "Building Go libvicore.so C-Shared library"
)

add_custom_target(libvicore_target DEPENDS ${CMAKE_CURRENT_BINARY_DIR}/libvicore.so)

add_library(bamboo-viet MODULE addon.cpp)
add_dependencies(bamboo-viet libvicore_target)

target_include_directories(bamboo-viet PRIVATE ${CMAKE_CURRENT_BINARY_DIR})
target_link_libraries(bamboo-viet PRIVATE Fcitx5::Core Fcitx5::Utils ${CMAKE_CURRENT_BINARY_DIR}/libvicore.so)
set_target_properties(bamboo-viet PROPERTIES PREFIX "")

install(TARGETS bamboo-viet DESTINATION "${CMAKE_INSTALL_LIBDIR}/fcitx5")
install(FILES ${CMAKE_CURRENT_BINARY_DIR}/libvicore.so DESTINATION "${CMAKE_INSTALL_LIBDIR}")

fcitx5_translate_desktop_file(bamboo-viet.conf.in bamboo-viet.conf)
fcitx5_translate_desktop_file(bamboo-viet-addon.conf.in bamboo-viet-addon.conf)

install(FILES "${CMAKE_CURRENT_BINARY_DIR}/bamboo-viet.conf" DESTINATION "${FCITX_INSTALL_PKGDATADIR}/inputmethod")
install(FILES "${CMAKE_CURRENT_BINARY_DIR}/bamboo-viet-addon.conf" DESTINATION "${FCITX_INSTALL_PKGDATADIR}/addon")
