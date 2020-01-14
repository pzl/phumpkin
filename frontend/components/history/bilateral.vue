<template>
	<simple-sliders :sliders="sliders" />
</template>

<script>
import SimpleSliders from '~/components/history/simpleSliders'

export default {
	props: {
		r: {},
		g: {},
		b: {},
		x: {},
		y: {},
		version: {},
	},
	data() {
		return {
			sliders: [
				{title: "radius", value: this.x, min: 1, max: 30, format: v => v.toFixed(1)},
				{title: "r", value: this.r, min: 0, max: 0.1, format: v => v.toFixed(4)},
				{title: "g", value: this.g, min: 0, max: 0.1, format: v => v.toFixed(4)},
				{title: "b", value: this.b, min: 0, max: 0.1, format: v => v.toFixed(4)},
			]
		}
	},
	components: { SimpleSliders },
}
</script>