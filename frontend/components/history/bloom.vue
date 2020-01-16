<template>
	<simple-sliders :sliders="sliders" />
</template>

<script>
import SimpleSliders from '~/components/history/simpleSliders'

export default {
	props: {
		size: {},
		threshold: {},
		strength: {},
		version: {},
	},
	data() {
		return {
			sliders: [
				{title: "size", value: this.size, min: 0, max: 100, format: v => v.toFixed(0)+'%'},
				{title: "threshold", value: this.threshold, min: 0, max: 100, format: v => v.toFixed(0)+'%'},
				{title: "strength", value: this.strength, min: 0, max: 100, format: v => v.toFixed(0)+'%'},
			]
		}
	},
	components: { SimpleSliders },
}
</script>