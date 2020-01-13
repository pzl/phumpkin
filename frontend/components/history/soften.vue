<template>
	<simple-sliders :sliders="sliders" />
</template>

<script>
import SimpleSliders from '~/components/history/simpleSliders'

export default {
	props: {
		amount: {},
		brightness: {},
		saturation: {},
		size: {},
		version: {},
	},
	data() {
		return {
			sliders: [
				{title: "size", value: this.size, min: 0, max: 100, format: v => v.toFixed(0)+'%'},
				{title: "sat", value: this.saturation, min: 0, max: 100, format: v => v.toFixed(0)+'%'},
				{title: "bright", value: this.brightness, min: -2, max: 2, format: v => v.toFixed(2)+'EV'},
				{title: "mix", value: this.amount, min: 0, max: 100, format: v => v.toFixed(0)+'%'},
			]
		}
	},
	components: { SimpleSliders },
}
</script>