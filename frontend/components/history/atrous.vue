<template>
	<v-sheet>
		<div class="eq-graph">
			<svg viewBox="0 0 100 60">
				<rect x="0" y="0" width="100" height="60" :fill="fill" />
				<text x="0" y="6" :fill="stroke">Luma</text>
				<text x="30" y="6" fill="rgba(149,96,42,0.8)">Chroma</text>
				<text x="70" y="6" fill="rgba(40,106,172,0.8)">Edges</text>
				<path fill="rgba(149,96,42,0.3)" stroke="rgba(149,96,42,0.5)" stroke-width="0.5"
					:d="toPoints(chrominance, chrominance_noise)" />
				<path fill="rgba(255,255,255,0.2)" :stroke="stroke" stroke-width="0.5"
					:d="toPoints(luminance, luminance_noise)" />
				<path fill="none" stroke="rgba(40,106,172,0.4)" :d="toPoints(sharpness)" />
			</svg>
		</div>
	</v-sheet>
</template>

<script>
import { line, area, curveNatural } from 'd3-shape'

export default {
	props: {
		octaves: {}, // idk what this does
		luminance: {},
		luminance_noise: {},
		chrominance: {},
		chrominance_noise: {},
		sharpness: {},
		version: {},
	},
	data() {
		return {
		}
	},
	methods: {
		toPoints(top, bottom) {
			if (bottom) {
				const combined = []
				for (const i in top) {
					combined.push({
						x: top[i].x*100,
						y0: 60-top[i].y*60,
						y1: 60-bottom[i].y*60,
					})
				}
				return area().x(d => d.x).y0(d => d.y0).y1(d => d.y1).curve(curveNatural)(combined)
			}

			return line().x(d=>d.x*100).y(d=>60-d.y*60).curve(curveNatural)(top)
		},
	},
	computed: {
		fill() { return this.$vuetify.theme.dark ? "#4c4c4c" : "#ddd" },
		stroke() { return this.$vuetify.theme.dark ? "#b3b3b3" : "#333" },
	},
}
</script>


<style>

.eq-graph text {
	font-size: 7px;
	font-family: sans-serif;
}

</style>