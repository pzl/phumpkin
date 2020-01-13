<template>
	<v-sheet>
		<div class="monoc-graph">
			<svg :viewBox="'0 0 '+boxsize+' '+boxsize">
				<rect
					v-for="(b,i) in blocks" :key="i"
					:width="square_size" :height="square_size"
					:x="b.x" :y="b.y"
					:fill="b.fill"
					/>
				<line fill="none" stroke="rgba(170,170,170,1)" :x1="wx" :x2="bx" :y1="wy" :y2="by" />
				<circle :cx="bx" :cy="by" r="1.3" fill="black" />
				<circle :cx="wx" :cy="wy" r="1.3" fill="white" />
			</svg>
		</div>
		<simple-sliders :sliders="sliders" />
	</v-sheet>
</template>

<script>
import { lab } from 'd3-color'
import SimpleSliders from '~/components/history/simpleSliders'

const nBoxes = 8
const ccmax = 40.0

export default {
	props: {
		hi_a: {}, // a = x
		hi_b: {}, // b = y
		low_a: {},
		low_b: {},
		saturation: {},
		version: {},
	},
	data() {
		return {
			sliders: [
				{title: "sat", value: this.saturation, min: -3, max: 3, format: v => v.toFixed(2)},
			],
			square_size: 13, // px
		}
	},
	methods: {
		to_pt(ab, invert) {
			if (invert) {
				return 0.5 * (this.boxsize + -1 * this.boxsize * ab / ccmax)
			}
			return 0.5 * (this.boxsize + this.boxsize * ab / ccmax)
		}
	},
	computed: {
		boxsize() { return this.square_size*nBoxes },
		wx() { return this.to_pt(this.hi_a) },
		wy(){ return this.to_pt(this.hi_b, true) },
		bx(){ return this.to_pt(this.low_a) },
		by(){ return this.to_pt(this.low_b, true) },
		blocks() {
			const blocks = []
			for (let i =0; i<nBoxes; i++) {
				for (let j=0; j<nBoxes; j++) {
					const l = 53.390011
					const a = this.saturation * (l * 0.05 * ccmax * (i/(nBoxes-1.0) - 0.5))
					const b = this.saturation * (l * 0.05 * ccmax * (j/(nBoxes-1.0) - 0.5))
					const c = lab(l,a,b)
					blocks.push({
						x: i*this.square_size,
						y: (this.boxsize-this.square_size) - j*this.square_size, // invert Y
						fill: c.toString(),
					})
				}
			}
			return blocks
		}
	},
	components: { SimpleSliders },
}
</script>