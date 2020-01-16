<template>
	<v-sheet>
		<p class="ma-0">{{ camera }}</p>
		<p class="ma-0">{{ lens }}</p>
		<v-row justify="space-between">
			<span>{{ focal }}mm</span>
			<span>f/ {{ aperture }}</span>
			<span>d {{ distance }}</span>
		</v-row>
		<v-row justify="space-between">
			<span>corrections</span>
			<span>{{ corr }}</span>
		</v-row>
		<v-row justify="space-between">
			<span>geometry</span>
			<span>{{ target_geo }}</span>
		</v-row>
		<v-row justify="space-between">
			<span>mode</span>
			<span>{{ inverse }}</span>
		</v-row>
		<simple-sliders class="mt-3" :sliders="sliders" />
		<v-row justify="space-between">
			<span>tca override</span>
			<span>{{ tca_override }}</span>
		</v-row>
		<v-row justify="space-between">
			<span>modified</span>
			<span>{{ modified }}</span>
		</v-row>
	</v-sheet>
</template>

<script>
import SimpleSliders from '~/components/history/simpleSliders'

export default {
	props: {
		corrections: {},
		inverse: {}, // is this mode?
		scale: {},
		crop: {},
		focal: {},
		aperture: {},
		distance: {},
		target_geo: {},
		camera: {},
		lens: {},
		tca_override: {}, // or is this mode
		tca_r: {},
		tca_b: {},
		modified: {},
		version: {},
	},
	data() {
		return {
			sliders: [
				{title: "scale", value: this.scale, min: 0.1, max: 2, format: v => v.toFixed(3)},
				{title: "tca red", value: this.tca_r, min: 0.99, max: 1.01, format: v => v.toFixed(5)},
				{title: "tca blue", value: this.tca_b, min: 0.99, max: 1.01, format: v => v.toFixed(5)},
			]
		}
	},
	computed: {
		corr() {
			return this.corrections // @todo convert to string
		}
	},
	components: { SimpleSliders },
}
</script>