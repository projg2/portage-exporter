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

#ifndef _VDB_HH
#define _VDB_HH

#include <filesystem>
#include <vector>

#include <prometheus/family.h>
#include <prometheus/gauge.h>
#include <prometheus/registry.h>

class vdb_collector {
  private:
    prometheus::Family<prometheus::Gauge>& m_vdb;
    std::vector<prometheus::Gauge*> m_metrics;

  public:
    vdb_collector(prometheus::Registry& registry);

    void collectInstalled(const std::filesystem::path& vdbPath);
};

#endif  // _VDB_HH
