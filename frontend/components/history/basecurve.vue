<template>
	<v-sheet>
		<div class="lowlight-graph">
			<svg :viewBox="'0 0 '+width+' '+height">
				<rect x="0" y="0" :width="width" :height="height" :fill="fill" />
				<line 
					v-for="i in 4" :key="'v'+i"
					:x1="scx(i/4)" :x2="scx(i/4)" y1="0" :y2="height"
					:stroke="gridlines"
					/>
				<line 
					v-for="i in 4" :key="'h'+i"
					x1="0" :x2="width" :y1="scy(i/4)" :y2="scy(i/4)"
					:stroke="gridlines"
					/>
				<line x1="0" :y1="scy(0)" :x2="scx(1)" :y2="scy(1)" :stroke="gridlines" stroke-dasharray="4" />
				<path fill="none" :stroke="stroke" :d="linedata" />
				<circle
					v-for="(p,i) in pts" :key="'p'+i"
					:cx="scx(p.x)" :cy="scy(p.y)" r="1.4"
					:stroke="stroke" stroke-width="0.6"
					:fill="fill" 
					>
					<title>{{ p.x }}, {{ p.y }}</title>
				</circle>
			</svg>
		</div>
		<v-row no-gutters justify="space-between">
			<span style="width: 40%">scale</span>
			<v-select v-model="scale" :items="scale_types" dense hide-details height="20" class="caption" style="width: 60%" />
		</v-row>
		<v-row no-gutters justify="space-between">
			<span>preserve colors</span>
			<span>{{preserve_color}}</span>
		</v-row>
		<v-row no-gutters justify="space-between">
			<span>fusion</span>
			<span>{{exposure_fusion == 0 ? 'none' : exposure_fusion}}</span>
		</v-row>
	</v-sheet>
</template>

<script>
import { line, area, curveCardinal, curveCatmullRom, curveMonotoneX } from 'd3-shape'

export default {
	props: {
		curve: {},
		curve_type: {},
		exposure_bias: {},
		exposure_fusion: {},
		exposure_stops: {},
		n_nodes: {},
		preserve_color: {},
		version: {},
	},
	data() {
		return {
			height: 100,
			width: 100,
			scale: 'linear',
			scale_types: ['linear','logarithm']
		}
	},
	methods: {
		scx(x) {
			const adjust = this.scale === 'linear' ?  (x) => x : this.log
			return adjust(x)*this.width
		},
		scy(y) {
			const adjust = this.scale === 'linear' ?  (x) => x : this.log
			return this.height - adjust(y)*this.height
		},
		log(x) {
			const base = 64
			return Math.log(x * (base-1) + 1) / Math.log(base)
		},
	},
	computed: {
		pts() {
			return this.curve.filter((d,i)=>i<this.n_nodes)
		},
		linedata() {
			let curve = curveCardinal
			switch (this.curve_type) {
				case "Cubic spline": curve = curveCardinal; break;
				case "Catmull-Rom": curve = curveCatmullRom; break;
				case "Monotone Hermite": curve = curveMonotoneX; break;
			}
			return line()
					.x(d=>this.scx(d.x))
					.y(d=>this.scy(d.y))
					.curve(curve)(this.pts)
		},
		fill() { return this.$vuetify.theme.dark ? "#4c4c4c" : "#ddd" },
		stroke() { return this.$vuetify.theme.dark ? "#b3b3b3" : "#333" },
		gridlines() { return this.$vuetify.theme.dark ? '#444' : '#d0d0d0' }
	},
}
</script>


<style>

.lowlight-graph text {
	font-size: 7px;
	font-family: sans-serif;
}

</style>