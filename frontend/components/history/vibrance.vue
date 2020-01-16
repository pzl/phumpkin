<template>
	<simple-sliders :sliders="sliders" />
</template>

<script>
import SimpleSliders from '~/components/history/simpleSliders'

export default {
	props: {
		amount: {},
		version: {},
	},
	data() {
		return {
			sliders: [
				{title: "amount", value: this.amount, min: 0, max: 100, format: v => v+'%' },
			]
		}
	},
	components: { SimpleSliders}
}
</script>