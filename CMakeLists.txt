cmake_minimum_required(VERSION 3.12 FATAL_ERROR)

project(portage-exporter
    VERSION 0.1.0
    DESCRIPTION "Prometheus exporter for Gentoo Portage"
    HOMEPAGE_URL "https://github.com/projg2/portage-exporter"
    LANGUAGES CXX
)

find_package(prometheus-cpp CONFIG REQUIRED)
include(GNUInstallDirs)

add_executable(${PROJECT_NAME} src/main.cc src/vdb.cc)
target_link_libraries(${PROJECT_NAME} PRIVATE prometheus-cpp::pull)
set_property(TARGET ${PROJECT_NAME} PROPERTY CXX_STANDARD 17)
install(TARGETS ${PROJECT_NAME} DESTINATION ${CMAKE_INSTALL_BINDIR})
