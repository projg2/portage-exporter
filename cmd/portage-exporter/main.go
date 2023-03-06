// portage-exporter - Prometheus exporter for Gentoo Portage
// Copyright (C) 2023 Arthur Zamarin
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/projg2/portage-exporter/portage/vdb"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()

	vdb.Collector(getEnvDuration("VDB_UPDATE_INTERVAL", 5*time.Minute), getEnvString("VDB_PATH", "/var/db/pkg"), ctx)

	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":2112", nil)
}

func getEnvDuration(key string, def time.Duration) time.Duration {
	val := os.Getenv(key)
	if val == "" {
		return def
	}

	dur, err := time.ParseDuration(val)
	if err != nil {
		return def
	}

	return dur
}

func getEnvString(key string, def string) string {
	val := os.Getenv(key)
	if val == "" {
		return def
	}
	return val
}
