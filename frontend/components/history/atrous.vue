<template>
	<v-sheet>
		<div class="eq-graph">
			<svg viewBox="0 0 100 100">
				<rect x="0" y="0" width="100" height="100" :fill="fill" />
				<text x="0" y="8">Luma</text>
				<path fill="rgba(255,255,255,0.3)" :stroke="stroke"
					:d="toPoints(luminance, luminance_noise)" />
			</svg>
			<svg viewBox="0 0 100 100">
				<rect x="0" y="0" width="100" height="100" :fill="fill" />
				<text x="0" y="8">Chroma</text>
				<path fill="rgba(255,255,255,0.3)" :stroke="stroke"
					:d="toPoints(chrominance, chrominance_noise)" />
			</svg>
			<svg viewBox="0 0 100 100">
				<rect x="0" y="0" width="100" height="100" :fill="fill" />
				<text x="0" y="8">Edges</text>
				<path fill="none" :stroke="stroke" :d="toPoints(sharpness)" />
			</svg>
		</div>
	</v-sheet>
</template>

<script>
import SimpleSliders from '~/components/history/simpleSliders'
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
						y0: 100-top[i].y*100,
						y1: 100-bottom[i].y*100,
					})
				}
				return area().x(d => d.x).y0(d => d.y0).y1(d => d.y1).curve(curveNatural)(combined)
			}

			return line().x(d=>d.x*100).y(d=>100-d.y*100).curve(curveNatural)(top)
		},
	},
	computed: {
		fill() { return this.$vuetify.theme.dark ? "#4c4c4c" : "#ddd" },
		stroke() { return this.$vuetify.theme.dark ? "#b3b3b3" : "#333" },
	},
	components: { SimpleSliders },
}
</script>


<style>

.eq-graph text {
	font-size: 9px;
	font-family: sans-serif;
	fill: #ccc;
}

</style>