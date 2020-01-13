<template>
	<simple-sliders :sliders="sliders" />
</template>

<script>
import SimpleSliders from '~/components/history/simpleSliders'

export default {
	props: {
		offset_a: {}, // ?
		offset_b: {}, // ?
		steep_a: {},
		steep_b: {},
		unbound: {}, // ?
		version: {},
	},
	data() {
		return {
			sliders: [
				{title: "gr v mag", value: this.steep_a, min: 0, max: 5, format: v => v.toFixed(2)},
				{title: "bl v y", value: this.steep_b, min: 0, max: 5, format: v => v.toFixed(2)},
			]
		}
	},
	components: { SimpleSliders },
}
</script>