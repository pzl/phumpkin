<template>
	<v-sheet>
		<simple-sliders :sliders="sliders" />
		<v-row no-gutters justify="space-between">
			<p>soften with</p>
			<p>{{algo}}</p>
		</v-row>
	</v-sheet>
</template>

<script>
import SimpleSliders from '~/components/history/simpleSliders'

export default {
	props: {
		algo: {},
		brightness: {},
		contrast: {},
		order: {},
		radius: {},
		saturation: {},
		unbound: {},
		version: {},
	},
	data() {
		return {
			sliders: [
				{title: "radius", value: this.radius, min: 0, max: 500, format: v => v.toFixed(2)},
				{title: "contrast", value: this.contrast, min: -3, max: 3, format: v => v.toFixed(2)},
				{title: "brightness", value: this.brightness, min: -3, max: 3, format: v => v.toFixed(2)},
				{title: "saturation", value: this.saturation, min: -3, max: 3, format: v => v.toFixed(2)},
			]
		}
	},
	components: { SimpleSliders },
}
</script>