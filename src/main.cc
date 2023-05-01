// portage-exporter - Prometheus exporter for Gentoo Portage
// Copyright (C) 2023 Arthur Zamarin
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

#include <CivetServer.h>
#include <prometheus/exposer.h>
#include <prometheus/registry.h>

#include <chrono>
#include <cstdlib>
#include <filesystem>
#include <iostream>
#include <memory>
#include <string>
#include <thread>

#include "vdb.hh"

using namespace std::literals::chrono_literals;

static std::string getenv_str(const char* env_var, const char* default_value) {
    const char* value = std::getenv(env_var);
    return (value && *value) ? value : default_value;
}

static std::chrono::seconds getenv_seconds(const char* env_var,
                                           std::chrono::seconds default_value) {
    const char* value = std::getenv(env_var);
    if (!value) {
        return default_value;
    }
    try {
        int seconds = std::atoi(value);
        return seconds > 0 ? std::chrono::seconds(seconds) : default_value;
    } catch (...) {
        return default_value;
    }
}

int main() {
    const auto vdbPath =
        std::filesystem::path(getenv_str("VDB_PATH", "/var/db/pkg"));
    const auto interval = getenv_seconds("VDB_UPDATE_INTERVAL", 5min);
    auto address = getenv_str("SERVE_ADDRESS", ":2112");
    if (address[0] == ':') {
        address = "0.0.0.0" + address;
    }

    try {
        auto registry = std::make_shared<prometheus::Registry>();
        prometheus::Exposer exposer{address};
        vdb_collector vdb{*registry};
        exposer.RegisterCollectable(registry,
                                    getenv_str("SERVE_PATH", "/metrics"));

        for (;;) {
            try {
                vdb.collectInstalled(vdbPath);
            } catch (const std::exception& e) {
                std::cerr << "failed to collect installed: " << e.what()
                          << std::endl;
            }
            std::this_thread::sleep_for(interval);
        }
    } catch (const CivetException& e) {
        std::cerr << "failed to init HTTP server: " << e.what() << std::endl;
        return 1;
    }
    return 0;
}
