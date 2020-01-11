<template>
	<simple-sliders :sliders="sliders" />
</template>

<script>
import SimpleSliders from '~/components/history/simpleSliders'

export default {
	props: {
		amount: {},
		radius: {},
		threshold: {},
		version: {},
	},
	data() {
		return {
			sliders: [
				{title: "radius", value: this.radius, min: 0, max: 8, format: v => v.toFixed(2)},
				{title: "amount", value: this.amount, min: 0, max: 2, format: v => v.toFixed(2)},
				{title: "threshold", value: this.threshold, min: 0, max: 100, format: v => v.toFixed(2)},
			]
		}
	},
	components: { SimpleSliders },
}
</script>