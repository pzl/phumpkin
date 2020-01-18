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
				<circle :cx="cx" :cy="cy" :r="boxsize*0.22*size" fill="none" stroke="rgba(255,255,255,0.4)" stroke-width="1" />
			</svg>
		</div>
		<simple-sliders :sliders="sliders" />
	</v-sheet>
</template>

<script>
import { lab } from 'd3-color'
import SimpleSliders from '~/components/history/simpleSliders'

const nBoxes = 8
const square_size = 13

export default {
	props: {
		a: {},
		b: {},
		highlights: {},
		size: {},
		version: {},
	},
	data() {
		return {
			sliders: [
				{title: "hi", value: this.highlights, min: 0, max: 1, format: v => v.toFixed(2)},
			],
			square_size: square_size,
			boxsize: square_size * nBoxes,
		}
	},
	methods: {
		clamp: (v, h, l) => v > l ? v < h ? v : h : l,
		filter_booster(ai, bi, a, b, size) {
			return Math.exp(
				-1 * this.clamp( (Math.pow(ai-a,2) + Math.pow(bi-b,2))/(2*size) , 1, 0)
			)
		}
	},
	computed: {
		cx() {
			return this.a/256.0 * this.boxsize + this.boxsize/2.0
		},
		cy() {
			return this.boxsize - (this.b/256.0 * this.boxsize + this.boxsize/2.0)
		},
		blocks() {
			const blocks = []
			for (let i =0; i<nBoxes; i++) {
				for (let j=0; j<nBoxes; j++) {
					// see monochrome.c :: dt_iop_monochrome_draw
					let l = 53.390011
					const a = 256.0 * ( i/(nBoxes-1.0) - 0.5)
					const b = 256.0 * ( j/(nBoxes-1.0) - 0.5)
					let f = this.filter_booster(a,b,this.a,this.b, 40*40*this.size*this.size)
					l *= f*f;
					const c = lab(l,a,b)
					blocks.push({
						x: i*square_size,
						y: (this.boxsize-square_size) - j*square_size, // invert Y
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