<template>
	<simple-sliders :sliders="sliders" />
</template>

<script>
import SimpleSliders from '~/components/history/simpleSliders'

export default {
	props: {
		contrast: {},
		f_size: {},
		version: {},
	},
	data() {
		return {
			sliders: [
				{title: "contrast cmpr", value: this.contrast, min: 1, max: 5, format: v => v.toFixed(3)},
				{title: "spatial extent", value: this.f_size, min: 0, max: 100, format: v => v.toFixed(0)+'%'},
			]
		}
	},
	components: { SimpleSliders },
}
</script>