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
		<svg id="svgcanvas" width="960" height="500">
			<defs>
			<marker id="marker-circle" markerWidth="2" markerHeight="2" refX="1" refY="1">
				<circle cx="1" cy="1" r="0.75"/>
			</marker>
			</defs>
		</svg>
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

// Figure out the function to parse each component of a data point
// assumes data looks like [x0, x1, x2, ...]
var inferScaleDomain = function(data) {
	var tmParsers = {
		"%m%d %H:%M:%S"			: d3.timeParse("%m%d %H:%M:%S"),
		"%Y%m%d %H:%M:%S"		: d3.timeParse("%Y%m%d %H:%M:%S"),
		"%Y/%m/%d %H:%M:%S" : d3.timeParse("%Y/%m/%d %H:%M:%S"),
		"%Y-%m-%d %H:%M:%S" : d3.timeParse("%Y-%m-%d %H:%M:%S"),
		"%Y-%m-%d"					: d3.timeParse("%Y-%m-%d"),
		"%Y/%m/%d"					: d3.timeParse("%Y/%m/%d"),
		"%B %d, %Y"					: d3.timeParse("%B %d, %Y"),
	};
	var freq = {};
	data.forEach(function(p) {
		Object.keys(tmParsers).forEach(function(fmt) {
			var v = tmParsers[fmt](p);
			if (v !== null) { freq[fmt] += 1; }
		});
	});
	if (Object.keys(freq).length == 0) {
		return {
			parser: function(p) {return +p;},
			scale: d3.scaleLinear,
		};
	}
	var fmt = Object.keys(freq).reduce(function(a, b) {
		return freq[a] > freq[b] ? a : b;
	});
	return {
		parser: tmParsers[fmt],
		scale: d3.scaleTime,
	};
};

var lineChart = function(config, scale) {
	var margin = {t: 20, r: 80, b: 30, l: 50},
			svg = d3.select("#svgcanvas"),
			width = svg.attr("width") - margin.l - margin.r,
			height = svg.attr("height") - margin.t - margin.b,
			frame = svg.append("g")
				.attr("transform", "translate(" + margin.l + "," + margin.t + ")");
	// generate SVG path, dataXform maps data points to target domain
	var line = function(dataXform) {
			return d3.line()
			.curve(d3.curveMonotoneX)
			.x(dataXform.x)
			.y(dataXform.y)
	};
	// use data to reset axes
	var updateAxes = function(scale) {
		frame.selectAll("g.axis").remove();
		frame.append("g")
				.attr("class", "y axis")
				.call(d3.axisLeft(scale.y)
				.tickSizeInner(-width)
				.tickSizeOuter(0));
		frame.append("g")
				.attr("class", "x axis")
				.attr("transform", "translate(0," + height + ")")
				.call(d3.axisBottom(scale.x));
	};
	// function to call on data updates
	// data looks like [[x0, y0], [x1, y1], ...]
	var plotter = function(data) {
		var dataX = data.map(function(p){return p[0];});
		var domainX = inferScaleDomain(dataX);
		var extents = {
			x: d3.extent(data, function(p) {return domainX.parser(p[0]);}),
			y: d3.extent(data, function(p) {return +p[1];}),
		};
		var scale = {
			x: domainX.scale().range([0, width]).domain(extents.x),
			y: d3.scaleLinear().range([height, 0]).domain(extents.y),
		};
		var xform = {
			x: function(d) {return scale.x(domainX.parser(d[0]));},
			y: function(d) {return scale.y(+d[1]);},
		};
		updateAxes(scale);
		frame.selectAll("path.dataline").remove();
		frame.append("path")
			.datum(data)
			.attr("class", "line dataline")
			.attr("d", line(xform));
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
