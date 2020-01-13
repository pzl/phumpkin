<template>
	<v-sheet>
		<div class="lowlight-graph">
			<svg :viewBox="'0 0 '+width+' '+height">
				<rect x="0" y="0" :width="width" :height="height" :fill="fill" />
				<line 
					v-for="i in 8" :key="'v'+i"
					:x1="width/8*i" :x2="width/8*i" y1="0" :y2="height"
					:stroke="gridlines"
					/>
				<line 
					v-for="i in 8" :key="'h'+i"
					x1="0" :x2="width" :y1="height/8*i" :y2="height/8*i"
					:stroke="gridlines"
					/>
				<text :x="width/2" y="6" :fill="stroke" text-anchor="middle">day vision</text>
				<text :x="width/2" :y="height-3" :fill="stroke" text-anchor="middle">night vision</text>
				<text :x="-1*height/2" y="6" :fill="stroke" transform="rotate(-90)" text-anchor="middle">dark</text>
				<text :x="-1*height/2" :y="width-3" :fill="stroke" transform="rotate(-90)" text-anchor="middle">bright</text>
				<path fill="none" :stroke="stroke" :d="linedata" />
				<circle
					v-for="(p,i) in pts" :key="'p'+i"
					:cx="p[0]*width" :cy="height-p[1]*height" r="1.4"
					:stroke="stroke" stroke-width="0.6"
					:fill="fill" 
					>
					<title>{{ p[0] }}, {{ p[1] }}</title>
				</circle>
			</svg>
		</div>
		<simple-sliders :sliders="sliders" />
	</v-sheet>
</template>

<script>
import SimpleSliders from '~/components/history/simpleSliders'
import { line, area, curveNatural } from 'd3-shape'

export default {
	props: {
		blueness: {},
		transition_x: {},
		transition_y: {},
		version: {},
	},
	data() {
		return {
			height: 100,
			width: 135,
			sliders: [
				{title: "blue shift", value: this.blueness, min: 0, max: 100, format: v => v.toFixed(0)+'%'},
			]
		}
	},
	methods: {
	},
	computed: {
		pts() {
			const unified = []
			for (let i = 0; i < this.transition_x.length; i++) {
				unified.push([this.transition_x[i],this.transition_y[i]])
			}
			return unified
		},
		linedata() {
			return line().x(d=>d[0]*this.width).y(d=>this.height - d[1]*this.height).curve(curveNatural)(this.pts)
		},
		fill() { return this.$vuetify.theme.dark ? "#4c4c4c" : "#ddd" },
		stroke() { return this.$vuetify.theme.dark ? "#b3b3b3" : "#333" },
		gridlines() { return this.$vuetify.theme.dark ? '#444' : '#d0d0d0' }
	},
	components: { SimpleSliders },
}
</script>


<style>

.lowlight-graph text {
	font-size: 7px;
	font-family: sans-serif;
}

</style>