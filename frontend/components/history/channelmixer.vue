<template>
	<v-sheet>
		<v-row no-gutters justify="space-between">
			<span style="width: 40%">dest</span>
			<v-select v-model="dest" :items="destinations" dense hide-details height="20" class="caption" style="width: 60%" />
		</v-row>
		<simple-sliders :sliders="sliders" />
	</v-sheet>
</template>

<script>
import SimpleSliders from '~/components/history/simpleSliders'

export default {
	props: {
		hue: {},
		saturation: {},
		lightness: {},
		red: {},
		green: {},
		blue: {},
		grey: {},
		version: {},
	},
	data() {
		return {
			sliders: [
				{title: "red", value: this.hue.r, min: -2, max: 2, format: v => v.toFixed(3)},
				{title: "green", value: this.hue.g, min: -2, max: 2, format: v => v.toFixed(3)},
				{title: "blue", value: this.hue.b, min: -2, max: 2, format: v => v.toFixed(3)},
			],
			destinations: ['hue', 'saturation', 'lightness', 'red', 'green', 'blue', 'grey'],
			dest: 'hue',
		}
	},
	watch: {
		dest(v) {
			this.sliders[0].value = this[v].r
			this.sliders[0].value = this[v].g
			this.sliders[0].value = this[v].b
		}
	},
	components: { SimpleSliders },
}
</script>