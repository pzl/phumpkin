<template>
	<v-sheet>
		<v-select v-model="viewing" :items="view_by" dense hide-details height="20" class="caption" />
		<div class="cz-graph" :style="{ height: height+'px', width: width+'px' }">
			<canvas :width="width" :height="height" ref="cv"></canvas>
			<svg :viewBox="'0 0 '+width+' '+height">
				<path
					v-for="(c,i) in curve" :key="'curve'+i" 
					:stroke="selected(i) ? 'rgba(255,255,255,0.8)' : 'rgba(255,255,255,0.2)'" stroke-width="1.5" fill="none"
					:d="lines(i)"
				/>
				<circle
					v-for="(p,i) in pts(view_idx)" :key="'p'+i"
					:cx="p.x*width" :cy="height-p.y*height" r="2"
					stroke="rgba(255,255,255,0.8)" stroke-width="1"
					fill="none" 
					>
					<title>{{ p.x }}, {{ p.y }}</title>
				</circle>
			</svg>
		</div>
		<v-row no-gutters justify="space-between" class="caption">
			<span>select by</span>
			<span>{{ channel }}</span>
		</v-row>
		<v-row no-gutters justify="space-between" class="caption mb-4">
			<span>process mode</span>
			<span>{{ mode }}</span>
		</v-row>
		<simple-sliders :sliders="sliders" />
		<v-row no-gutters justify="space-between" class="caption">
			<span>interp</span>
			<span>{{ curve_type[view_idx] }}</span>
		</v-row>
	</v-sheet>
</template>

<script>
import SimpleSliders from '~/components/history/simpleSliders'
import { scaleLinear } from 'd3-scale'
import { interpolateHclLong } from 'd3-interpolate'
import { line, area, curveCardinal, curveCatmullRom, curveMonotoneX } from 'd3-shape'
import { color, rgb, lch } from 'd3-color'

export default {
	props: {
		mode: {}, // smooth v strong
		channel: {}, // x-axis, essentially
		curve: {}, // [3]
		curve_type: {},
		n_nodes: {}, // [3]
		strength: {},
		version: {},
		sm: {}
	},
	data() {
		return {
			width: 180,
			height: 180,
			sliders: [
				{title: "mix", value: this.strength, min: -200, max: 200, format: v => v.toFixed(2)+'%'},
			],
			viewing: 'lightness',
			view_by: ['lightness', 'saturation', 'hue'],
		}
	},
	methods: {
		drawBG() {
			const norm_c = 128 * Math.sqrt(2)
			const sample_color = lch(rgb(0,77,179))

			// set x-axis scale ranges based on operating channel
			let start_color
			let end_color
			switch (this.channel) {
				case 'Hue':
					start_color = lch(50, norm_c/2, 0)
					end_color = start_color.copy({h: 360 })
					break
				case 'Saturation':
					start_color = lch(50, 1, sample_color.h)
					end_color = start_color.copy({ c: 100 })
					break
				case 'Lightness':
					start_color = lch(2, norm_c/2, sample_color.h)
					end_color = start_color.copy({ l: 90 })
					break;
			}


			// set y-axis scale ranges based on viewed channel
			let sy = scaleLinear().domain([0, this.height])
			switch (this.viewing) {
				case 'lightness': sy = sy.range([80,1]); break; // lightness 80->1
				case 'saturation': sy = sy.range([80,5]); break; // saturation 80->1
				case 'hue': sy = sy.range([-180,180]); break; // hue, with 0 at midpoint
			}
			const scale = scaleLinear()
							.domain([0,this.width])
							.interpolate(interpolateHclLong)


			const data = this.ctx.createImageData(this.width, this.height)
			let i=0;
			for (let y=0; y<this.height; y++) {
				let yv = sy(y) // get y value for this row
				const change = (() => { // how the start/end values for x axis should behave with this Y val
					switch (this.viewing) {
						case 'lightness': return c => c.copy({l:yv}); break;
						case 'saturation': return c => c.copy({c: yv}); break;
						case 'hue': return  c => {
							let nc = c.copy()
							nc.h -= yv
							return nc
						}
					}
				})()
				// create row scale
				let sx = scale.range([change(start_color),change(end_color)])
				for (let x=0; x<this.width; x++) {
					let c = color(sx(x))
					data.data[i++] = c.r
					data.data[i++] = c.g
					data.data[i++] = c.b
					data.data[i++] = 255
				}
			}
			this.ctx.putImageData(data, 0, 0)
		},
		pts(idx) {
			return this.curve[idx].filter((d,i) => i< this.n_nodes[idx])
		},
		crvtype(idx) {
			switch (this.curve_type[idx]) {
				case "Cubic spline": return curveCardinal
				case "Catmull-Rom": return curveCatmullRom
				case "Monotone Hermite": return curveMonotoneX
			}
			return curveCardinal		
		},
		lines(idx) {
			return line().x(d=>d.x*this.width).y(d=>this.height-d.y*this.height).curve(this.crvtype(idx))(this.pts(idx))
		},
		selected(idx) {
			return this.view_idx == idx
		}
	},
	computed: {
		fill() { return this.$vuetify.theme.dark ? "#4c4c4c" : "#ddd" },
		stroke() { return this.$vuetify.theme.dark ? "#b3b3b3" : "#333" },
		canvas() { return this.$refs.cv },
		ctx() { return this.canvas.getContext('2d', { alpha: false }) },
		view_idx() { return this.view_by.indexOf(this.viewing) },
	},
	watch: {
		viewing() {
			this.drawBG()
		},
	},
	mounted() {
		this.drawBG()
	},
	components: { SimpleSliders },
}
</script>


<style>

.cz-graph {
	position: relative;
}

.cz-graph canvas, .cz-graph svg {
	position: absolute;
	top: 0;
}

</style>