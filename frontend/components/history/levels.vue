<template>
	<v-sheet>
		<v-row no-gutters justify="space-between">
			<p>Mode</p>
			<p>{{mode}}</p>
		</v-row>
		<simple-sliders v-if="mode=='automatic'" :sliders="sliders" />
		<div v-else class="levels-graph">
			<svg :viewBox="'0 0 '+width+' '+height">
				<rect x="0" y="0" :width="width" :height="height" :fill="fill" />
				<line v-for="(x,i) in gridlines" :key="'grid'+i" :x1="x" :x2="x" y1="0" :y2="height" :stroke="grid" stroke-width="0.5" />
				<path :d="hist" fill="rgba(50,50,50,0.4)" stroke="none" />
				<line v-for="(x,i) in lines" :key="'line'+i" :x1="x" :x2="x" y1="0" :y2="height" :stroke="stroke" stroke-width="0.5" />
			</svg>
		</div>
	</v-sheet>
</template>

<script>
import SimpleSliders from '~/components/history/simpleSliders'
import { lab } from 'd3-color'
import { histogram, extent } from 'd3-array'
import { area, curveNatural } from 'd3-shape'
import { scaleLinear } from 'd3-scale'

export default {
	props: {
		mode: {},
		levels: {},
		percentiles: {},
		version: {},
		sm: {},
	},
	data() {
		return {
			sliders: [
				{title: "black", value: this.percentiles[0], min: 0, max: 100, format: v => v.toFixed(1)+'%'},
				{title: "grey", value: this.percentiles[1], min: 0, max: 100, format: v => v.toFixed(1)+'%'},
				{title: "white", value: this.percentiles[2], min: 0, max: 100, format: v => v.toFixed(1)+'%'},
			],
			width: 100,
			height: 60,
			l: []
		}
	},
	methods: {
		px() {
			const canvas = document.createElement('canvas')
			canvas.width = this.sm.width
			canvas.height = this.sm.height
			const ctx = canvas.getContext('2d')
			const img = new Image()
			img.crossOrigin = ''

			img.onload = () => {
				ctx.drawImage(img, 0, 0)
				let data = ctx.getImageData(0,0,canvas.width,canvas.height).data

				// return every 4th pixel. A pixel is RGBA so every 4*4=16 entries.
				// need pixels 0,1,2,3 for RGBA for the 4th pixel
				const spread = 1
				data = data.filter((d,i) => i%(spread*4) < 4)

				const l = []
				for (let i = 0; i<data.length; i += 4) {
					l.push(lab(data[i],data[i+1],data[i+2]))
				}
				this.l = l.map(d=>d.l)
			}
			img.src = this.sm.url
		},
	},
	computed: {
		fill() { return this.$vuetify.theme.dark ? "#4c4c4c" : "#ddd" },
		grid() { return this.$vuetify.theme.dark ? "#3a3a3a" : "#ccc" },
		stroke() { return this.$vuetify.theme.dark ? "#b3b3b3" : "#333" },
		gridlines() {
			return [1,2,3].map(d => Math.floor(d*this.width/4)+0.5 )
		},
		lines() {
			return this.levels.map(d=>Math.floor(d*this.width)+0.5)
		},
		hist() {
			const compress = 2 // create 256/4 bins, to just combine a bit, graph doesn't need to be high res
			const thresh = []
			for (let i=0; i<256; i+=compress) {
				thresh.push(i) 
			}
			const hv = histogram().thresholds(thresh)(this.l).map(d=>Math.log(1+d.length))
			const scY = scaleLinear().domain(extent(hv)).range([this.height,0])
			const scX = scaleLinear().domain([0,256/compress]).range([this.levels[0]*this.width,this.levels[2]*this.width])
			return area().x((d,i)=>scX(i)).y0(d=>this.height).y1(d=>scY(d))(hv)
		}
	},
	mounted() {
		this.px()
	},
	components: { SimpleSliders },
}
</script>