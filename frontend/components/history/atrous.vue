<template>
	<v-sheet>
		<div class="eq-graph">
			<svg viewBox="0 0 100 100">
				<rect x="0" y="0" width="100" height="100" :fill="fill" />
				<text x="0" y="8">Luma</text>
				<polygon fill="rgba(255,255,255,0.3)" :stroke="stroke"
					:points="toPoints(luminance) + ' ' +toPoints(luminance_noise.slice().reverse())" />
			</svg>
			<svg viewBox="0 0 100 100">
				<rect x="0" y="0" width="100" height="100" :fill="fill" />
				<text x="0" y="8">Chroma</text>
				<polygon fill="rgba(255,255,255,0.3)" :stroke="stroke"
					:points="toPoints(chrominance) + ' ' +toPoints(chrominance_noise.slice().reverse())" />
			</svg>
			<svg viewBox="0 0 100 100">
				<rect x="0" y="0" width="100" height="100" :fill="fill" />
				<text x="0" y="8">Edges</text>
				<polyline fill="none" :stroke="stroke" :points="toPoints(sharpness)" />
			</svg>
		</div>
	</v-sheet>
</template>

<script>
import SimpleSliders from '~/components/history/simpleSliders'

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
		toPoints(data) {
			return data.map(p => (p.x*100)+","+(100-p.y*100)).join(" ")
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