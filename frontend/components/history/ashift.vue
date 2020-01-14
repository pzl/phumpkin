<template>
	<v-sheet>
		<simple-sliders :sliders="sliders" />
		<v-row no-gutters justify="space-between" class="pa-0 ma-0 caption">
			<span class="pa-0 ma-0">Auto Crop</span>
			<span class="pa-0 ma-0">{{crop}}</span>
		</v-row>
		<v-row no-gutters justify="space-between" class="pa-0 ma-0 caption">
			<span class="pa-0 ma-0">Lens model</span>
			<span class="pa-0 ma-0">{{mode}}</span>
		</v-row>
		<simple-sliders v-if="mode == 'specific'" :sliders="spec_sliders" />
	</v-sheet>
</template>

<script>
import SimpleSliders from '~/components/history/simpleSliders'

export default {
	props: {
		rotation: {},
		lens_shift_v: {},
		lens_shift_h: {},
		shear: {},
		f_length: {},
		crop_factor: {},
		ortho_corr: {},
		aspect: {},
		mode: {},
		toggle: {},
		crop: {},
		cl: {},
		cr: {},
		ct: {},
		cb: {},
		version: {},
	},
	data() {
		return {
			sliders: [
				{title: "rot", value: this.rotation, min: -10, max: 10, format: v => v.toFixed(2)+'°'},
				{title: "lens shf V", value: this.lens_shift_v, min: -1, max: 1, format: v => v.toFixed(3)},
				{title: "lens shf H", value: this.lens_shift_h, min: -1, max: 1, format: v => v.toFixed(3)},
				{title: "shear", value: this.shear, min: -0.2, max: 0.2, format: v => v.toFixed(3)},
			],
			spec_sliders: [
				{title: "focal", value: this.f_length, min: 10, max: 1000, format: v => v.toFixed(0)+'mm'},
				{title: "crop factor", value: this.crop_factor, min: 1, max: 2, format: v => v.toFixed(2)},
				{title: "aspect adj", value: this.aspect, min: 0.5, max: 2, format: v => v.toFixed(2)},
			]
		}
	},
	components: { SimpleSliders },
}
</script>