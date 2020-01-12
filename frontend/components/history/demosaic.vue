<template>
	<v-sheet>
		<v-row no-gutters justify="space-between">
			<p>Method</p>
			<p>{{method}}</p>
		</v-row>
		<simple-sliders :sliders="sliders" />
		<v-row no-gutters justify="space-between">
			<p>Color smoothing</p>
			<p>{{ color_smoothing }}</p>
		</v-row>
		<v-row no-gutters justify="space-between">
			<p>match greens</p>
			<p>{{ green_eq }}</p>
		</v-row>
	</v-sheet>
</template>

<script>
import SimpleSliders from '~/components/history/simpleSliders'

export default {
	props: {
		method: {},
		median_threshold: {},
		color_smoothing: {},
		green_eq: {},
		version: {},
	},
	data() {
		return {
			sliders: [
				{title: "edge thresh", value: this.median_threshold, min: 0, max: 1, format: v => v.toFixed(3)},
			]
		}
	},
	components: { SimpleSliders },
}
</script>