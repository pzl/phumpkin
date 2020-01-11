<template>
	<v-sheet>
		<v-row no-gutters justify="space-between">
			<p>Mode</p>
			<p>{{mode}}</p>
		</v-row>
		<simple-sliders :sliders="sliders" />
	</v-sheet>
</template>

<script>
import SimpleSliders from '~/components/history/simpleSliders'

export default {
	props: {
		mode: {},
		detail: {},
		sigma_r: {},
		sigma_s: {},
		midtone: {},
		version: {},
	},
	data() {
		return {
			sliders: [
				{title: "detail", value: this.detail*100, min: 0, max: 500, format: v => v.toFixed(0)+'%'},
				{title: "highlights", value: this.sigma_r*100, min: 0, max: 200, format: v => v.toFixed(0)+'%'},
				{title: "shadows", value: this.sigma_s*100, min: 0, max: 200, format: v => v.toFixed(0)+'%'},
				{title: "midtone range", value: this.detail, min: 0, max: 1, format: v => v.toFixed(3)},
			]
		}
	},
	components: { SimpleSliders },
}
</script>