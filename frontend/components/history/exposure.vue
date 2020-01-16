<template>
	<v-sheet>
		<v-row no-gutters justify="space-between">
			<p>Mode</p>
			<p>{{mode}}</p>
		</v-row>
		<simple-sliders :sliders="sliders" />
	</v-sheet>
</template>

<script>
import SimpleSliders from '~/components/history/simpleSliders'

export default {
	props: {
		mode: {},
		black: {},                  // always used
		exposure: {},               // used if mode == manual
		deflicker_target_level: {}, // used if mode == automatic
		deflicker_percentile: {},   // used if mode == automatic
		version: {},
	},
	data() {
		return {
			sliders: [
				{title: "exposure", value: this.exposure, min: -3, max: 3, format: v => v.toFixed(2)+'EV'},
				//{title: "clip", value: this.amount, min: 0, max: 0.2, format: v => v.toFixed(4)},
				{title: "black", value: this.black, min: -0.1, max: 0.1, format: v => v.toFixed(4)},
			]
		}
	},
	components: { SimpleSliders },
}
</script>