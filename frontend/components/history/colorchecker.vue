<template>
	<v-sheet>
		<div class="monoc-graph">
			<div class="d-flex mb-4" style="flex-wrap: wrap;">
				<div
					v-for="(b,i) in blocks" :key="i"
					style="height: 26px; width: 26px; cursor: pointer;"
					:style="{
						backgroundColor: b.rgb,
						border: highlighted === i ? '1px solid red' : b.same ? 'none' : '1px solid white',
					}"
					@click="select(i)"
					></div>
			</div>
		</div>
		<simple-sliders :sliders="sliders" />
	</v-sheet>
</template>

<script>
import { lab } from 'd3-color'
import SimpleSliders from '~/components/history/simpleSliders'

const nBoxes = 8

export default {
	props: {
		n_patches: {},
		source: {},
		target: {},
		version: {},
		sm: {},
	},
	data() {
		return {
			sliders: [
				{title: "light", value: 0, min: -100, max: 200, format: v => v.toFixed(2)},
				{title: "gr/red", value: 0, min: -256, max: 256, format: v => v.toFixed(2)},
				{title: "bl/yl", value: 0, min: -256, max: 256, format: v => v.toFixed(2)},
				{title: "sat", value: 0, min: -128, max: 128, format: v => v.toFixed(2)},
			],
			highlighted: null,
		}
	},
	methods: {
		select(i) {
			this.highlighted = i
			this.sliders[0].value = this.target[i].l - this.source[i].l
			this.sliders[1].value = this.target[i].a - this.source[i].a
			this.sliders[2].value = this.target[i].b - this.source[i].b
			// idk what to do about saturation
		}
	},
	computed: {
		boxsize() {
			return this.square_size*nBoxes
		},
		blocks() {
			return this.source
					.filter((d,i)=>i<this.n_patches)
					.map((d,i)=>{
						const src = lab(d.l,d.a,d.b)
						const dest = lab(this.target[i].l, this.target[i].a, this.target[i].b)
						return {
							src: src,
							rgb: src.toString(),
							dest: dest,
							same: src.toString() == dest.toString(),
						}
					})
		}
	},
	components: { SimpleSliders },
}
</script>