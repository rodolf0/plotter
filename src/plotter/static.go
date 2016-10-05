// vi:syntax=javascript
package main

import (
	"html/template"
)

var GraphHtml = template.Must(template.New("").Parse(`
<!DOCTYPE html>
<html>
	<head>
		<meta charset="utf-8" />
		<title>Plotter</title>
		<script src="https://ajax.googleapis.com/ajax/libs/jquery/3.1.1/jquery.min.js"></script>
		<script src="//d3js.org/d3.v4.min.js"></script>
		<style>
		.line {
			fill: none;
			stroke: steelblue;
			stroke-width: 1.5px;
			marker-mid: url(#marker-circle);
			marker-end: url(#marker-circle);
			marker-start: url(#marker-circle);
		}
		.tick line{
			opacity: 0.1;
		}
		</style>
	</head>
	<body>
		<h2 id="workspace">Workspace</h2>
		<svg id="svgcanvas" width="960" height="500"></svg>
		<script>

var setupWS = function(msgCallback) {
	// setup polling for new data to the server
	var ws = new WebSocket("{{.Wsaddr}}");
	ws.onerror = function(errevent) {
		var t = $('#workspace').html();
		$('#workspace').html(t + " - ERROR");
		console.log('Websocket error: ' + errevent);
	};
	ws.onclose = function(ev) {
		ws = null;
		var t = $('#workspace').html();
		$('#workspace').html(t + " - LOST");
		console.log('Websocket closed');
	};
	ws.onmessage = function(ev) {
		var data = JSON.parse(ev.data);
		// set workspace title
		if ("workspace" in data) {
			$('#workspace').html(data["workspace"]);
			return;
		}
		msgCallback(data.Cells);
	}
};

var tmParsers = {
	"%m%d %H:%M:%S"			:  d3.timeParse("%m%d %H:%M:%S"),
	"%Y%m%d %H:%M:%S"		:  d3.timeParse("%Y%m%d %H:%M:%S"),
	"%Y/%m/%d %H:%M:%S" :  d3.timeParse("%Y/%m/%d %H:%M:%S"),
	"%Y-%m-%d %H:%M:%S" :  d3.timeParse("%Y-%m-%d %H:%M:%S"),
	"%Y-%m-%d"					:  d3.timeParse("%Y-%m-%d"),
	"%Y/%m/%d"					:  d3.timeParse("%Y/%m/%d"),
	"%B %d, %Y"					:  d3.timeParse("%B %d, %Y"),
};

// assumes data looks like
// [[x0, y0], [x1, y1], ...]
var detectAxis = function(data, axis) {
	var freq = {};
	data.forEach(function(point) {
		Object.keys(tmParsers).forEach(function(fmt) {
			if (tmParsers[fmt](point[axis]) !== null) {
				freq[fmt] += 1;
			}
		});
	});
	if (Object.keys(freq).length != 0) {
		return Object.keys(freq).reduce(function(a, b) {
			return freq[a] > freq[b] ? a : b;
		});
	}
	return null;
};

var lineChart = function() {
	var svg = d3.select("#svgcanvas"),
			margin = {top: 20, right: 80, bottom: 30, left: 50},
			width = svg.attr("width") - margin.left - margin.right,
			height = svg.attr("height") - margin.top - margin.bottom;
	var x = d3.scaleLinear().range([0, width]),
			y = d3.scaleLinear().range([height, 0]);

	svg.append("defs").append("marker")
		.attr("id", "marker-circle")
		.attr("markerWidth", "2").attr("markerHeight", "2")
		.attr("refX", "1").attr("refY", "1")
		.append("circle")
		.attr("cx", "1").attr("cy", "1").attr("r", "0.75")
		.attr("class", "marker");

	// setup svg frame of reference
	var frame = svg.append("g")
			.attr("transform", "translate(" + margin.left + "," + margin.top + ")");
	// line setup
	var line = function(xparser) {
			return d3.line()
			.curve(d3.curveMonotoneX)
			.x(function(d) { return x(xparser(d)); })
			.y(function(d) { return y(+d[1]); });
	};
	// setup initial axes
	var updateAxes = function(data, xparser) {
		frame.selectAll("g.axis").remove();
		x.domain(d3.extent(data, xparser));
		y.domain(d3.extent(data, function(d) { return +d[1]; }));
		frame.append("g")
				.attr("class", "y axis")
				.call(d3.axisLeft(y).tickSizeInner(-width).tickSizeOuter(0));
		frame.append("g")
				.attr("class", "x axis")
				.attr("transform", "translate(0," + height + ")")
				.call(d3.axisBottom(x));
	};
	updateAxes([[0, 0], [10, 10]], function(d){return +d[0]});
	// plotter function to call on new data
	var plotter = function(data) {
		var xfmt = detectAxis(data, 0);
		var xparser = function(d) { return +d[0]; };
		x = d3.scaleLinear().range([0, width]);
		if (xfmt !== null) {
			x = d3.scaleTime().range([0, width]);
			xparser = function(d) { return tmParsers[xfmt](d[0]); };
		}

		updateAxes(data, xparser);
		// https://github.com/d3/d3-selection#joining-data
		frame.selectAll("path.dataline").remove();
		frame.append("path")
			.datum(data)
			.attr("class", "line dataline")
			.attr("d", line(xparser));
	};
	return plotter;
};

$(function() {
	var plotter = lineChart();
	setupWS(plotter);
});

		</script>
	</body>
</html>
`))
