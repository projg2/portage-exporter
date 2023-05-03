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

#include "vdb.hh"

#include <fstream>
#include <functional>
#include <iostream>
#include <regex>

static const std::regex package_name_re{"(.+)-([0-9]+[^-]*)(?:-(r[0-9]+))?",
                                        std::regex::optimize};

vdb_collector::vdb_collector(prometheus::Registry& registry)
    : m_vdb(prometheus::BuildGauge()
                .Name("portage_package")
                .Help("Installed packages")
                .Register(registry)) {}

static std::string read_file(std::filesystem::path path) {
    std::ifstream file{path, std::ios::in | std::ios::binary};
    if (!file.is_open())
        return {};
    return {std::istreambuf_iterator<char>(file),
            std::istreambuf_iterator<char>()};
}

static std::string trim(std::string str) {
    str.erase(std::find_if(str.rbegin(), str.rend(),
                           [](char ch) { return !std::isspace(ch); })
                  .base(),
              str.end());
    str.erase(str.begin(), std::find_if(str.begin(), str.end(), [](char ch) {
                  return !std::isspace(ch);
              }));
    return str;
}

void vdb_collector::collectInstalled(const std::filesystem::path& vdbPath) {
    for (auto metric : m_metrics) {
        m_vdb.Remove(metric);
    }
    m_metrics.clear();

    for (auto const& category : std::filesystem::directory_iterator{vdbPath}) {
        try {
            if (!category.is_directory()) {
                continue;
            }
            for (auto const& package :
                 std::filesystem::directory_iterator{category}) {
                std::smatch matches;
                const auto PF = package.path().filename().string();
                if (std::regex_match(PF, matches, package_name_re) &&
                    matches.size() >= 3) {
                    const auto PN = matches.str(1);
                    const auto PV = matches.str(2);
                    const auto PR = matches.size() > 3 ? matches.str(3) : "";
                    const auto PVR = PR.empty() ? PV : PV + "-" + PR;

                    auto& m = m_vdb.Add(
                        {{"CATEGORY", category.path().filename()},
                         {"P", PN + "-" + PV},
                         {"PF", PF},
                         {"PN", PN},
                         {"PR", PR},
                         {"PV", PV},
                         {"PVR", PVR},
                         {"repository",
                          trim(read_file(package.path() / "repository"))},
                         {"SLOT", trim(read_file(package.path() / "SLOT"))}});
                    m.Set(1);
                    m_metrics.push_back(&m);
                }
            }
        } catch (const std::filesystem::filesystem_error& err) {
            std::cerr << "failed to read " << category.path() << ": "
                      << err.what() << std::endl;
        }
    }
}
