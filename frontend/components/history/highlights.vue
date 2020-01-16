<template>
	<v-sheet>
		<v-row no-gutters justify="space-between">
			<p>method</p>
			<p>{{ mode }}</p>
		</v-row>
		<simple-sliders :sliders="sliders" />
	</v-sheet>
</template>

<script>
import SimpleSliders from '~/components/history/simpleSliders'

export default {
	props: {
		mode: {},
		clip: {},
		version: {},
	},
	data() {
		return {
			sliders: [
				{title: "clip thrs", value: this.clip, min: 0, max: 2, format: v => v.toFixed(3)},
			]
		}
	},
	components: { SimpleSliders },
}
</script>