<template>
	<v-sheet class="history-item">
		<simple-sliders :sliders="sliders" />
		<v-row no-gutters justify="space-between">
			<p>intent</p>
			<p>{{ interpolator }}</p>
		</v-row>
		<v-row no-gutters justify="space-between">
			<v-checkbox dense class="pa-0 ma-0" :value="preserve_color" label="preserve chrominance" readonly disabled />
		</v-row>
	</v-sheet>
</template>

<script>
import SimpleSliders from '~/components/history/simpleSliders'

export default {
	props: {
		balance: {},
		black_point_source: {},
		black_point_target: {},
		contrast: {},
		global_saturation: {},
		grey_point_source: {},
		grey_point_target: {},
		interpolator: {},
		latitude: {},
		latitude_stops: {},
		output_power: {},
		preserve_color: {},
		saturation: {},
		security: {},
		white_point_source: {},
		white_point_target: {},
		version: {},
	},
	data() {
		return {
			sliders: [
				{title: "grey", value: this.grey_point_source, min: 0.1, max: 36, format: v => v.toFixed(2)+'%'},
				{title: "white", value: this.white_point_source, min: 1.14, max: 5.2, format: v => v.toFixed(2)+'EV'},
				{title: "black", value: this.black_point_source, min: -15, max: 3, format: v => v.toFixed(2)+'EV'},
				{title: "safety", value: this.security, min: -50, max: 50, format: v => v.toFixed(2)+'%'},
				{title: "contrast", value: this.contrast, min: 1, max: 2, format: v => v.toFixed(3)},
				{title: "latitude", value: this.latitude, min: 2, max: 8, format: v => v.toFixed(2)+'EV'},
				{title: "shd/hi bal", value: this.balance, min: -50, max: 50, format: v => v.toFixed(2)+'%'},
				{title: "global sat", value: this.global_saturation, min: 2, max: 8, format: v => v.toFixed(2)+'%'},
				{title: "extr lum sat", value: this.recalc_extreme_lum(this.saturation), min: 0, max: 200, format: v => v.toFixed(2)+'%'},
			]
		}
	},
	methods: {
		// see filmic.c::saturation_callback
		recalc_extreme_lum(d) { return ((Math.exp(d / 100 * Math.log(10) )-1)*100/9) }
	},
	computed: {
		extreme_lum() { return this.recalc_extreme_lum(this.saturation) }
	},
	components: { SimpleSliders },
}
</script>

<style>
.history-item .v-input--selection-controls.v-input .v-label {
	font-size: 13px;
}
</style>