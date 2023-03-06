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

package vdb

import (
	"context"
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	promInstalled = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "portage_package",
		Help: "Installed packages",
	}, []string{"category", "pkgver", "slot", "repository"})

	promDuration = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "portage_installed_duration",
		Help: "Duration of the last collection of installed packages.",
	})
)

func collectInstalled(vdb string) {
	timer := prometheus.NewTimer(prometheus.ObserverFunc(promDuration.Set))
	defer timer.ObserveDuration()

	cats, err := os.ReadDir(vdb)
	if err != nil {
		fmt.Println("Failed to read directory", vdb, "err:", err)
		return
	}

	promInstalled.Reset()
	for _, cat := range cats {
		if !cat.IsDir() {
			continue
		}

		catPath := path.Join(vdb, cat.Name())
		packages, err := os.ReadDir(catPath)
		if err != nil {
			fmt.Println("Failed to read directory", catPath, "err:", err)
			continue
		}

		for _, pkg := range packages {
			if !pkg.IsDir() {
				continue
			}

			pkgPath := path.Join(catPath, pkg.Name())

			repo, err := os.ReadFile(path.Join(pkgPath, "repository"))
			if err != nil {
				fmt.Println("Failed to read repository file for", pkgPath, "err:", err)
				continue
			}

			slot, err := os.ReadFile(path.Join(pkgPath, "SLOT"))
			if err != nil {
				fmt.Println("Failed to read slot file for", pkgPath, "err:", err)
				continue
			}

			promInstalled.WithLabelValues(
				cat.Name(),
				pkg.Name(),
				strings.TrimSpace(string(slot)),
				strings.TrimSpace(string(repo)),
			).Set(1)
		}
	}
}

func Collector(duration time.Duration, vdbPath string, ctx context.Context) {
	go func() {
		ticker := time.NewTicker(duration)
		collectInstalled(vdbPath)
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				collectInstalled(vdbPath)
			}
		}
	}()
}
