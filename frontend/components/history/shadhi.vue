<template>
	<v-sheet>
		<simple-sliders :sliders="sliders" />
		<v-row no-gutters justify="space-between">
			<p>soften with</p>
			<p>{{algo}}</p>
		</v-row>
	</v-sheet>
</template>

<script>
import SimpleSliders from '~/components/history/simpleSliders'

export default {
	props: {
		algo: {},
		compress: {},
		flags: {}, // ?
		highlights: {},
		highlights_ccorrect: {},
		low_approximation: {}, // ?
		order: {}, // ?
		radius: {},
		shadows: {},
		shadows_ccorrect: {},
		whitepoint: {},
		version: {},
	},
	data() {
		return {
			sliders: [
				{title: "shadows", value: this.shadows, min: -100, max: 100, format: v => v.toFixed(2)},
				{title: "highlights", value: this.highlights, min: -100, max: 100, format: v => v.toFixed(2)},
				{title: "wht pt", value: this.whitepoint, min: -10, max: 10, format: v => v.toFixed(2)},
				{title: "radius", value: this.radius, min: 0, max: 500, format: v => v.toFixed(2)},
				{title: "compress", value: this.compress, min: 0, max: 100, format: v => v.toFixed(2)+'%'},
				{title: "shd col adj", value: this.shadows_ccorrect, min: 0, max: 100, format: v => v.toFixed(2)+'%'},
				{title: "hi col adj", value: this.highlights_ccorrect, min: 0, max: 100, format: v => v.toFixed(2)+'%'},
			]
		}
	},
	components: { SimpleSliders },
}
</script>