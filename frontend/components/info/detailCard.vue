<template>
	<v-card outlined  class="ma-2 mb-5 summary-card">
		<v-card-title><name :name="name" :exif="exif" /></v-card-title>
		<v-card-subtitle><rating :readonly="true" :value="meta.rating" /></v-card-subtitle>

		<v-tabs grow v-model="tab" height="20px">
			<v-tab style="min-width: 20px" class="pa-0" href="#basic"><v-icon small>mdi-calendar-blank</v-icon></v-tab>
			<v-tab style="min-width: 20px" class="pa-0" href="#shot-info"><v-icon small>mdi-camera-iris</v-icon></v-tab>
			<v-tab style="min-width: 20px" class="pa-0" href="#tags"><v-icon small>mdi-tag-heart</v-icon></v-tab>
			<v-tab style="min-width: 20px" class="pa-0" href="#edits" v-if="xmp.history && xmp.history.length"><v-icon small>mdi-image-edit</v-icon></v-tab>
			<v-tab style="min-width: 20px" class="pa-0" href="#copy"><v-icon small>mdi-copyright</v-icon></v-tab>
		</v-tabs>

		<v-card-text :class="{ 'pa-0': tab === 'edits' }">
			<v-tabs-items v-model="tab">
				<v-tab-item value="basic">
					<div>{{ sizeof }}</div>
					<div>{{ original.width }}x{{ original.height }}</div>
					<date :date="exif.DateTimeOriginal" :tz="exif.TimeZone" :offset="exif.OffsetTime" />
					<div class="loc" v-if="meta.loc"><v-icon small>mdi-map-marker</v-icon> {{ meta.loc.lat }}, {{ meta.loc.lon }}<span v-if="meta.loc.alt">, {{ meta.loc.alt }}</span></div>
				</v-tab-item>
				<v-tab-item value="shot-info">
					<v-row no-gutters justify="space-between">
						<v-col cols="1" v-if="metering">
							<v-icon :title="'Metering: '+exif.MeteringMode" small>{{ metering }}</v-icon>
						</v-col>
						<v-col cols="1" v-if="focus_icon">
							<v-icon :title="exif.AFAreaMode" small>{{ focus_icon }}</v-icon>
						</v-col>
						<v-col cols="1" v-if="flash">
							<v-icon :title="exif.Flash" small>{{ flash }}</v-icon>
						</v-col>
						<v-col cols="1" v-if="exif.ImageStabilization">
							<v-icon :title="'IS: '+exif.ImageStabilization" small>mdi-vibrate{{ exif.ImageStabilization.includes("On") ? '' : '-off' }}</v-icon>
						</v-col>
						<v-col cols="1" v-if="temp">
							<v-icon :title="temp+ ' | '+tempF" small>mdi-thermometer</v-icon>
						</v-col>
						<v-col cols="1" v-if="exif.FacesDetected">
							<v-icon :title="exif.FacesDetected" small>mdi-face-recognition</v-icon>
						</v-col>
						<v-col cols="1"  v-if="exif.SelfTimer && exif.SelfTimer !== 'Off'">
							<v-icon :title="exif.SelfTimer" small>mdi-camera-timer</v-icon>
						</v-col>
						<v-col cols="1" v-if="batt">
							<v-icon :title="exif.BatteryLevel" small>{{ batt }}</v-icon>
						</v-col>
					</v-row>
					<div v-if="exif">
						<v-row dense justify="space-between">
							<v-col cols="3">f / {{ exif.Aperture }}</v-col>
							<v-col cols="auto">{{ exposure }}</v-col>
							<v-col cols="auto">ISO: {{ exif.ISO }}</v-col>
						</v-row>
						<div>Focal Length: {{exif.FocalLength}}</div>
						<div>Focus Mode: {{exif.FocusMode}}</div>
					</div>

					<div v-if="exif.LensID">{{exif.LensID}}</div>
				</v-tab-item>
				<v-tab-item value="tags">
					<v-row v-if="colors" dense>
						<v-col v-for="(c,i) in colors" :key="i">
							<v-chip :color="c.toLowerCase()" small></v-chip>
						</v-col>
					</v-row>
					<tag-crumbs v-if="xmp.tags && xmp.tags.length" :tags="xmp.tags" />
				</v-tab-item>
				<v-tab-item value="edits" v-if="xmp.history && xmp.history.length">
					<v-row no-gutters class="pa-2" justify="space-between">
						<h4>History</h4>
						<v-btn v-if="edits_open.length" @click="edits_open=[]" icon small title="hide all"><v-icon small>mdi-eye-off</v-icon></v-btn>
					</v-row>
					<v-expansion-panels accordion multiple hover v-model="edits_open">
						<v-expansion-panel v-for="(h,i) in xmp.history.slice().reverse()" :key="i" :readonly="!h.params">
							<v-expansion-panel-header class="pa-2" style="min-height: 35px" :hide-actions="!h.params">
								<!--
								<template v-slot:actions>
									<v-btn icon small><v-icon small>mdi-eye-settings-outline</v-icon></v-btn>
								</template>
								-->
								<div class="d-flex">
									<span :title="h.op_name" :style="{ textDecoration: h.enabled ? 'none' : 'line-through' }">{{ h.name }}</span>
									<template v-if="h.multi_name">
										<v-spacer />
										<span class="caption">{{ h.multi_name }}</span>
									</template>
								</div>
							</v-expansion-panel-header>
							<v-expansion-panel-content class="pt-2">
								<component v-if="h.op_name in $options.components" :is="h.op_name" v-bind="h.params" :version="h.mod_version" :sm="thumbs.small" />
								<pre v-else>{{ JSON.stringify(h.params,null,2) }}</pre>
							</v-expansion-panel-content>
						</v-expansion-panel>
					</v-expansion-panels>
				</v-tab-item>
				<v-tab-item value="copy">
					<div v-if="meta.creator">Creator: {{ meta.creator }}</div>
					<div v-if="meta.rights">Rights: {{ meta.rights }}</div>
				</v-tab-item>
			</v-tabs-items>
		</v-card-text>

		<v-card-actions>
			<v-menu v-model="show_meta" :close-on-content-click="false" offset-x>
				<template v-slot:activator="{ on }">
					<v-btn v-on="on" icon>
						<v-icon small>mdi-details</v-icon>
					</v-btn>
				</template>


				<v-card class="raw-popout">
					<v-card-text class="text--primary xmp-popout">
						<pre>{{ JSON.stringify({"meta":meta,"xmp":xmp,"exif":exif}, null, 2) }}</pre>
					</v-card-text>
				</v-card>
			</v-menu>
		</v-card-actions>

	</v-card>
</template>

<script>
import TagCrumbs from '~/components/tagCrumbs'
import Rating from '~/components/rating'
import Name from '~/components/info/name'
import Date from '~/components/info/date'
import ashift from '~/components/history/ashift'
import basecurve from '~/components/history/basecurve'
import bilateral from '~/components/history/bilateral'
import bloom from '~/components/history/bloom'
import cacorrect from '~/components/history/cacorrect'
import channelmixer from '~/components/history/channelmixer'
import colorchecker from '~/components/history/colorchecker'
import colorzones from '~/components/history/colorzones'
import sharpen from '~/components/history/sharpen'
import colisa from '~/components/history/colisa'
import vibrance from '~/components/history/vibrance'
import exposure from '~/components/history/exposure'
import bilat from '~/components/history/bilat'
import demosaic from '~/components/history/demosaic'
import lens from '~/components/history/lens'
import levels from '~/components/history/levels'
import atrous from '~/components/history/atrous'
import filmic from '~/components/history/filmic'
import soften from '~/components/history/soften'
import monochrome from '~/components/history/monochrome'
import colorcorrection from '~/components/history/colorcorrection'
import shadhi from '~/components/history/shadhi'
import colorcontrast from '~/components/history/colorcontrast'
import velvia from '~/components/history/velvia'
import highlights from '~/components/history/highlights'
import lowlight from '~/components/history/lowlight'
import tonemap from '~/components/history/tonemap'
import lowpass from '~/components/history/lowpass'

export default {
	props: {
		name: {},
		dir: {},
		size: {}, // in bytes
		xmp: {},
		exif: {},
		meta: {}, // a few pre-merged XMP/EXIF fields
		thumbs: {}, // full: { url: "...", width: n, height: n}
		original: {}, //{ url: "...", width: n, height: n}
	},
	data() {
		return {
			show_meta: false,
			edits_open: [],
			tab: null,
		}
	},
	computed: {
		sizeof() {
			const units = ["b","KB","MB","GB","TB"]
			let unit = 0
			let b = this.size

			while (b > 1024) {
				b = b/1024
				unit++
			}
			return b.toFixed(2)+" "+units[unit]
		},
		batt() {
			if (!this.exif || !this.exif.BatteryLevel) {
				return null
			}
			const lvl = parseInt(this.exif.BatteryLevel.replace('%',''), 10)
			return 'mdi-battery-'+Math.ceil(lvl / 10) * 10
		},
		metering() {
			if (!this.exif || !this.exif.MeteringMode) {
				return null
			}
			switch (this.exif.MeteringMode) {
				case 'Multi-segment': return 'mdi-camera-metering-matrix'
			}
			return null
		},
		flash() {
			if (!this.exif || !this.exif.Flash) {
				return null
			}
			if (this.exif.Flash.startsWith("Off")) {
				return 'mdi-flash-off'
			}
			return null
		},
		exposure() {
			if (!this.exif || !this.exif.ExposureTime) {
				return null
			}
			if ( !(""+this.exif.ExposureTime).includes('/') ) {
				return this.exif.ExposureTime + "s"
			}
			return this.exif.ExposureTime
		},
		temp() {
			if (!this.exif) {
				return null
			}
			return this.exif.AmbientTemperature ||
					this.exif.CameraTemperature ||
					this.exif.BatteryTemperature ||
					null
		},
		tempF() {
			if (!this.temp) {
				return nill
			}
			return Math.round(parseInt(this.temp.replace(/[^\d]+/g,'')) * 9/5 + 32) + "F"
		},
		focus_icon() {
			if (!this.exif || !this.exif.AFAreaMode) {
				return null
			}
			switch (this.exif.AFAreaMode) {
				case 'Flexible Spot': return 'mdi-image-filter-center-focus'
				case 'Manual': return 'mdi-target'
				case 'Tracking': return 'mdi-face-recognition'
				case 'Auto': return 'mdi-focus-auto' // 'Auto' may not be actual string
			}
			return 'mdi-focus-field'
		},
		colors() {
			if (!this.xmp || !this.xmp.color_labels || !this.xmp.color_labels.length) {
				return null
			}
			return this.xmp.color_labels.map(c => {
				switch (c) {
					case "0": return "Red"
					case "1": return "Yellow"
					case "2": return "Green"
					case "3": return "Blue"
					case "4": return "Purple"
					default: return "Black"
				}
			})
		},
	},
	components: {
		Rating,
		TagCrumbs,
		Name,
		Date,
		bloom,
		channelmixer,
		sharpen,
		colisa,
		vibrance,
		exposure,
		bilat,
		demosaic,
		levels,
		atrous,
		filmic,
		soften,
		monochrome,
		colorcorrection,
		shadhi,
		colorcontrast,
		velvia,
		highlights,
		lowlight,
		tonemap,
		lens,
		lowpass,
		ashift,
		basecurve,
		bilateral,
		cacorrect,
		colorchecker,
		colorzones,
	}
}
</script>


<style>
.raw-popout {
	max-height: 500px;
}

.xmp-popout {
	background-color: white;
}

.theme--dark .xmp-popout {
	background-color: #303030;
}

.summary-card {
	overflow-y: auto;
	max-height: 85%;
}
</style>