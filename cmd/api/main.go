package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math"
	"net/http"
	"time"

	"github.com/metalmatze/athene/slo"
	"github.com/prometheus/client_golang/api"
	prometheusv1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
)

func main() {
	prometheusURL := flag.String("prometheus.url", "http://localhost:9090", "The URL to the Prometheus to query.")
	flag.Parse()

	log.Println("Using Prometheus at", *prometheusURL)

	client, err := api.NewClient(api.Config{Address: *prometheusURL})
	if err != nil {
		log.Fatal(err)
	}
	promAPI := prometheusv1.NewAPI(client)

	objective := slo.Objective{
		Target: 0.90,
		Window: model.Duration(30 * 24 * time.Hour),
		Indicator: slo.Indicator{
			HTTP: &slo.HTTPIndicator{
				Metric:    "prometheus_http_requests_total",
				Selectors: slo.Selectors{`handler="/debug/*subpath"`},
			},
		},
	}

	http.HandleFunc("/objective.json", sloHandler(promAPI, objective))
	http.HandleFunc("/objective/valet.json", valetHandler(promAPI, objective))
	http.HandleFunc("/objective/errorbudget.svg", svgHandler(promAPI, objective))

	if err := http.ListenAndServe(":9099", nil); err != nil {
		log.Fatal(err)
	}
}

func sloHandler(api prometheusv1.API, slo slo.Objective) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		sloStatus, err := status(r.Context(), api, slo)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		bytes, err := json.Marshal(sloStatus)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Access-Control-Allow-Origin", "*")
		_, _ = w.Write(bytes)
	}
}

type SLOStatus struct {
	Objective    Objective          `json:"objective"`
	Availability AvailabilityStatus `json:"availability"`
	Budget       BudgetStatus       `json:"budget"`
}

type Objective struct {
	Target float64        `json:"target"`
	Window model.Duration `json:"window"`
}

type AvailabilityStatus struct {
	Percentage float64 `json:"percentage"`
	Total      float64 `json:"total"`
	Errors     float64 `json:"errors"`
}

type BudgetStatus struct {
	Total     float64 `json:"total"`
	Remaining float64 `json:"remaining"`
	Max       float64 `json:"max"`
}

func status(ctx context.Context, api prometheusv1.API, objective slo.Objective) (SLOStatus, error) {
	status := SLOStatus{
		Objective: Objective{
			Target: objective.Target,
			Window: objective.Window,
		},
	}

	value, _, err := api.Query(ctx, objective.QueryTotal(objective.Window), time.Now().Round(time.Minute))
	if err != nil {
		return SLOStatus{}, err
	}
	vec := value.(model.Vector)
	for _, v := range vec {
		status.Availability.Total = float64(v.Value)
	}

	value, _, err = api.Query(ctx, objective.QueryErrors(objective.Window), time.Now().Round(time.Minute))
	if err != nil {
		return status, err
	}
	for _, v := range value.(model.Vector) {
		status.Availability.Errors = float64(v.Value)
	}

	status.Availability.Percentage = 1 - (status.Availability.Errors / status.Availability.Total)

	status.Budget.Total = 1 - objective.Target
	status.Budget.Remaining = (status.Budget.Total - (status.Availability.Errors / status.Availability.Total)) / status.Budget.Total
	status.Budget.Max = status.Budget.Total * status.Availability.Total

	return status, nil
}

func valetHandler(api prometheusv1.API, objective slo.Objective) http.HandlerFunc {
	type valet struct {
		Window       model.Duration `json:"window"`
		Volume       *float64       `json:"volume"`
		Errors       *float64       `json:"errors"`
		Availability *float64       `json:"availability"`
		Budget       *float64       `json:"budget"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		var valets []valet

		for _, window := range []model.Duration{
			objective.Window,
			model.Duration(7 * 24 * time.Hour),
			model.Duration(24 * time.Hour),
			model.Duration(time.Hour),
		} {
			totalVector, _, err := api.Query(r.Context(), objective.QueryTotal(window), time.Now())
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			errorsVector, _, err := api.Query(r.Context(), objective.QueryErrors(window), time.Now())
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			var total, errors, availability, budget *float64

			if len(totalVector.(model.Vector)) == 1 {
				value := float64(totalVector.(model.Vector)[0].Value)
				total = &value
			}
			if len(errorsVector.(model.Vector)) == 1 {
				value := float64(errorsVector.(model.Vector)[0].Value)
				errors = &value
			}

			if total != nil && errors != nil {
				av := 1 - *errors / *total
				availability = &av

				bv := ((1 - objective.Target) - (1 - *availability)) / (1 - objective.Target)
				budget = &bv
			}

			valets = append(valets, valet{
				Window:       window,
				Volume:       total,
				Errors:       errors,
				Availability: availability,
				Budget:       budget,
			})
		}

		bytes, err := json.Marshal(valets)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Access-Control-Allow-Origin", "*")
		_, _ = w.Write(bytes)
	}
}
func svgHandler(api prometheusv1.API, objective slo.Objective) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now().Add(-1 * time.Duration(objective.Window)).UTC()
		end := time.Now().UTC()

		query := objective.QueryErrorBudget()
		log.Println(query)
		value, _, err := api.QueryRange(r.Context(), query, prometheusv1.Range{
			Start: start,
			End:   end,
			Step:  time.Hour,
		})
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		width := 1200.0
		height := 320.0
		padding := 50.0

		if len(value.(model.Matrix)) == 0 {
			http.Error(w, "no data", http.StatusNotFound)
			return
		}

		var min, max float64
		{
			vMin := math.MaxFloat64
			vMax := 0.0
			for _, sample := range value.(model.Matrix)[0].Values {
				v := float64(sample.Value)
				if v > vMax {
					vMax = v
				}
				if v < vMin {
					vMin = v
				}
			}
			min = math.Floor(vMin*10) / 10
			max = math.Ceil(vMax*10) / 10
		}

		graph := fmt.Sprintf(`<g transform="translate(%.f,%.f)">`, padding, padding)
		for _, l := range graphLines(value.(model.Matrix)[0].Values, start, end, width-2*padding, height-2*padding, min, max) {
			graph += l
		}
		graph += `</g>`

		days := int(time.Duration(objective.Window).Hours() / 24)
		firstMidnight := time.Date(start.Year(), start.Month(), start.Day()+1, 0, 0, 0, 0, time.UTC)

		xAxis := fmt.Sprintf(`<g transform="translate(0,%.f)" fill="none" font-size="10" font-family="sans-serif" text-anchor="middle">`, height-padding)
		xAxis += fmt.Sprintf(`<path stroke="currentColor" d="M %.f 0 H %.f"/>`, padding, width-padding)
		for i := 0; i < days; i++ {
			midnight := firstMidnight.Add(time.Duration(i*24) * time.Hour)
			percentage := float64(midnight.Unix()-start.Unix()) / float64(end.Unix()-start.Unix())
			x := padding + percentage*(width-2*padding)
			xAxis += fmt.Sprintf(`<g transform="translate(%.f, 0)">`, x)
			xAxis += `<line stroke="currentColor" y2="6"/>`
			xAxis += fmt.Sprintf(`<text fill="currentColor" y="9" dy="0.71em" transform="translate(-25,20) rotate(-45)">%s</text>`, midnight.Format("2006-01-02"))
			xAxis += `</g>`
		}
		xAxis += `</g>`

		const labelCount = 10.0
		i := 0.0
		steps := (max - min) / labelCount // 10 value labels from max to min
		yAxis := fmt.Sprintf(`<g transform="translate(%.f,%.f)" fill="none" font-size="10" font-family="sans-serif" text-anchor="end">`, padding, padding)
		for v := max; v >= min; v = v - steps {
			yAxis += fmt.Sprintf(`<g class="tick" opacity="1" transform="translate(0,%.f)">`, i*(height-2*padding))
			yAxis += `<line stroke="currentColor" x1="-8" x2="-2"></line>`
			yAxis += fmt.Sprintf(`<text fill="currentColor" x="-10" dy="0.32em">%.f%%</text>`, v*100)
			yAxis += `</g>`
			i = i + (1 / labelCount)
		}
		yAxis += `</g>`

		out := fmt.Sprintf(`<svg viewBox="0,0,%.f,%.f" fill="none" xmlns="http://www.w3.org/2000/svg" width="%.f" height="%.f">`, width, height, width, height)
		out += graph
		out += xAxis
		out += yAxis
		out += fmt.Sprintf("<!--\n%s\n-->", query)
		out += `</svg>`

		w.Header().Set("Content-Type", "image/svg+xml")
		_, _ = fmt.Fprintln(w, out)
	}
}

type line struct {
	positive bool
	points   []point
}

type point struct {
	x, y float64
}

func graphLines(samples []model.SamplePair, start, end time.Time, width, height, min, max float64) []string {
	var lines []line

	var l line
	for i, value := range samples {
		if i == 0 {
			l.positive = float64(value.Value) > 0
		}
		if float64(value.Value) < 0 && l.positive {
			lines = append(lines, l)
			l = line{positive: false}
		}
		if float64(value.Value) > 0 && !l.positive {
			lines = append(lines, l)
			l = line{positive: true}
		}

		x := width * float64(value.Timestamp.Unix()-start.Unix()) / float64(end.Unix()-start.Unix())
		y := height - height*relative(min, max, float64(value.Value))
		l.points = append(l.points, point{x, y})
	}
	lines = append(lines, l)

	zeroRelative := relative(min, max, 0)
	zero := height - height*zeroRelative

	var paths []string
	for _, l := range lines {
		var path string
		for i, p := range l.points {
			if i == 0 {
				path = fmt.Sprintf("M%.f %.f ", p.x, p.y)
			} else {
				path += fmt.Sprintf("L%.f %.f ", p.x, p.y)
			}
		}

		if l.positive {
			if zero > height {
				zero = height
			}
			paths = append(paths,
				fmt.Sprintf(`<path stroke="#2C9938" stroke-width="1.5" stroke-linejoin="round" stroke-linecap="round" fill="none" d="%s"/>`, path),
			)
			paths = append(paths,
				fmt.Sprintf(`<path fill="#2C9938" fill-opacity="0.1" d="%sV%.f H%.f V%.f"/>`, path, zero, l.points[0].x, l.points[0].y), // TODO Might panic when no points...
			)
		} else {
			paths = append(paths,
				fmt.Sprintf(`<path stroke="#e6522c" stroke-width="1.5" stroke-linejoin="round" stroke-linecap="round" fill="none" d="%s"/>`, path),
			)
			paths = append(paths,
				fmt.Sprintf(`<path fill="#e6522c" fill-opacity="0.1" d="%sV%.f H%.f V%.f"/>`, path, zero, l.points[0].x, l.points[0].y), // TODO Might panic when no points...
			)
		}
	}

	if zeroRelative >= 0 && zeroRelative <= 1 {
		// only append the zero line if it's actually visible
		paths = append(paths,
			fmt.Sprintf(`<path stroke="#e6522c" stroke-width="1" stroke-dasharray="20,5" fill="none" d="M0 %.f H%.f"/>`, zero, width),
		)
	}

	return paths
}

func relative(min, max, v float64) float64 {
	return (-1*min + v) / (max - min)
}
