<template>
	<simple-sliders :sliders="sliders" />
</template>

<script>
import SimpleSliders from '~/components/history/simpleSliders'

export default {
	props: {
		contrast: {},
		brightness: {},
		saturation: {},
		version: {},
	},
	data() {
		return {
			sliders: [
				{title: "con", value: this.contrast, min: -1, max: 1, format: v => v.toFixed(2)},
				{title: "bri", value: this.brightness, min: -1, max: 1, format: v => v.toFixed(2)},
				{title: "sat", value: this.saturation, min: -1, max: 1, format: v => v.toFixed(2)},
			]
		}
	},
	components: { SimpleSliders },
}
</script>