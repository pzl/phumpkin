<template>
	<v-sheet>
		<v-row no-gutters justify="space-between">
			<p>Mode</p>
			<p>{{mode}}</p>
		</v-row>



		<simple-sliders v-if="mode=='automatic'" :sliders="sliders" />
		<div v-else class="levels-graph">
			<svg viewBox="0 0 100 60">
				<rect x="0" y="0" width="100" height="60" :fill="fill" />
				<line x1="33.5" x2="33.5" y1="0" y2="60" :stroke="thirds" stroke-width="0.5" />
				<line x1="66.5" x2="66.5" y1="0" y2="60" :stroke="thirds" stroke-width="0.5" />
				<line :x1="levels[0]*100" :x2="levels[0]*100" y1="0" y2="60" :stroke="stroke" stroke-width="1" />
				<line :x1="levels[1]*100" :x2="levels[1]*100" y1="0" y2="60" :stroke="stroke" stroke-width="1" />
				<line :x1="levels[2]*100" :x2="levels[2]*100" y1="0" y2="60" :stroke="stroke" stroke-width="1" />
			</svg>
		</div>
	</v-sheet>
</template>

<script>
import SimpleSliders from '~/components/history/simpleSliders'

export default {
	props: {
		mode: {},
		levels: {},
		percentiles: {},
		version: {},
	},
	data() {
		return {
			sliders: [
				{title: "black", value: this.percentiles[0], min: 0, max: 100, format: v => v.toFixed(1)+'%'},
				{title: "grey", value: this.percentiles[1], min: 0, max: 100, format: v => v.toFixed(1)+'%'},
				{title: "white", value: this.percentiles[2], min: 0, max: 100, format: v => v.toFixed(1)+'%'},
			]
		}
	},
	computed: {
		fill() { return this.$vuetify.theme.dark ? "#4c4c4c" : "#ddd" },
		thirds() { return this.$vuetify.theme.dark ? "#3a3a3a" : "#ccc" },
		stroke() { return this.$vuetify.theme.dark ? "#b3b3b3" : "#333" },
	},
	components: { SimpleSliders },
}
</script>