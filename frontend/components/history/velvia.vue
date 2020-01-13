<template>
	<simple-sliders :sliders="sliders" />
</template>

<script>
import SimpleSliders from '~/components/history/simpleSliders'

export default {
	props: {
		bias: {},
		strength: {},
		version: {},
	},
	data() {
		return {
			sliders: [
				{title: "strength", value: this.strength, min: 0, max: 100, format: v => v.toFixed(0)+'%'},
				{title: "midtone bias", value: this.bias, min: 0, max: 1, format: v => v.toFixed(2)},
			]
		}
	},
	components: { SimpleSliders },
}
</script>