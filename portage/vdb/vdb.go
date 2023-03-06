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
	"regexp"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	portagePackageLabels = []string{
		// labels in capital letters are directly from PMS:
		// https://projects.gentoo.org/pms/8/pms.html#x1-10900011.1
		"CATEGORY",
		"P",
		"PF",
		"PN",
		"PR",
		"PV",
		"PVR",
		"repository",
		"SLOT",
	}
	promInstalled = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "portage_package",
		Help: "Installed packages",
	}, portagePackageLabels)

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

			pf := pkg.Name()
			regex := regexp.MustCompile(`^(?P<PN>.+)(?:-(?P<PV>\d+[^-]*))(?:-(?P<PR>r\d{1,}))?$`)
			match := regex.FindStringSubmatch(pf)
			matchMap := make(map[string]string)
			for i, name := range regex.SubexpNames() {
				if i > 0 && i <= len(match) {
					matchMap[name] = match[i]
				}
			}


			// fmt.Printf("%#v\n", r.FindStringSubmatch(pf))
			// fmt.Printf("%#v\n", r.SubexpNames())
			pn := matchMap["PN"]
			p := matchMap["PN"] + "-" + matchMap["PV"]
			pr := matchMap["PR"]
			pv := matchMap["PV"]
			pvr := ""
			if(pr != "") {
				pvr = matchMap["PV"] + "-" + pr
			} else {
				pvr = matchMap["PV"]
				pr = "r0"
			}

			labelValueMap := prometheus.Labels{
				"CATEGORY": cat.Name(),
				"P": p,
				"PF": pf,
				"PN": pn,
				"PR": pr,
				"PV": pv,
				"PVR": pvr,
				"repository": strings.TrimSpace(string(repo)),
				"SLOT": strings.TrimSpace(string(slot)),
			}
			if(len(labelValueMap) != len(portagePackageLabels)) {
				panic("Mismatch between portagePackageLabels and labelValueMap")
			}
			promInstalled.With(labelValueMap).Set(1)
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
