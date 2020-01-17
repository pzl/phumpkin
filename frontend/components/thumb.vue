<template>
	<v-hover v-slot:default="{ hover }" open-delay="50" close-delay="150">
		<v-card :class="{ 'selected': isSelected }" :raised="isSelected" class="my-1 mx-auto thumb-card" >
			<v-img
				class="thumby"
				@click="onClick"
				@error="error"
				:src="src"
				:lazy-src="thumbs['x-small'].url + '?purpose=lazysrc'"
				:aspect-ratio="thumbs['large'].width/thumbs['large'].height"
				:height="height"
				:width="width"
				max-width="100%"
				contain
				v-ripple
				>
				<!-- 				:srcset="`
				${thumbs['x-small'].url} ${thumbs['x-small'].width}w,
				${thumbs['small'].url} ${thumbs['small'].width}w,
				${thumbs['medium'].url} ${thumbs['medium'].width}w,
				${thumbs['large'].url} ${thumbs['large'].width}w,
				${thumbs['x-large'].url} ${thumbs['x-large'].width}w,
				`"
				-->
				<template v-slot:placeholder>
					<v-row class="fill-height ma-0" align="center" justify="center">
						<v-lazy> <!-- prevent all below-fold images from spiraling forever and wasting cpu -->
							<v-progress-circular indeterminate v-if="!hasErr" color="grey lighten-5"></v-progress-circular>
							<v-icon v-else class="error--text" large>mdi-alert-outline</v-icon>
						</v-lazy>
					</v-row>
				</template>
			</v-img>

			<v-icon v-if="isSelected" class="select-check" color="success" large>mdi-checkbox-marked-circle-outline</v-icon>

			<size-select :x="menu_x" :y="menu_y" :thumbs="thumbs" v-model="menu" />

			<v-expand-transition v-if="scale > 0.18">
				<v-container v-if="hover" class="transition-fast-in-fast-out black darken-2 v-card--reveal white--text hidden-sm-and-down" fluid>
					<v-row dense class="d-flex justify-space-between align-center">
						<div>{{ name }}</div>
						<rating :value="display_rating" @input="rate({ image: index, rating: $event })" />
					</v-row>
					<div class="d-flex align-center">
						<v-tooltip bottom v-if="meta.loc">
							<template v-slot:activator="{ on }">
								<v-icon dark x-small v-if="meta.loc" v-on="on">mdi-map-marker</v-icon>
							</template>
							{{ meta.loc.lat }}, {{ meta.loc.lon }}
						</v-tooltip>
						<tags :dark="true" :tags="tags" />
						<v-spacer />
						<v-btn icon dark small :href="original.url">
							<v-icon>mdi-download</v-icon>
						</v-btn>
						<v-btn icon dark small @click="showMenu">
							<v-icon>mdi-dots-vertical</v-icon>
						</v-btn>
					</div>
				</v-container>
			</v-expand-transition>
		</v-card>
	</v-hover>
</template>

<script>
import Rating from '~/components/rating'
import Tags from '~/components/tags'
import SizeSelect from '~/components/sizeSelect'
import { mapState, mapMutations, mapActions } from 'vuex'

export default {
	props: {
		index: {}, // not part of image, but it's place in the grid
		name: {},
		dir: {},
		size: {}, // in bytes
		rating: {},
		rotation: {}, // 0-7 int
		tags: {}, // array of strings
		xmp: {}, //
		meta: {},
		exif: {},
		thumbs: {}, // full: { url: "...", width: n, height: n}
		original: {}, //{ url: "...", width: n, height: n}
	},
	data() {
		return {
			hover_reject: false,
			menu: false,
			menu_x: 0,
			menu_y: 0,
			hasErr: false,
			srcOverride: false,
		}
	},
	computed: {
		isSelected(){
			return this.selected.includes(this.index)
		},
		display_rating() {
			return this.hover_reject ? 0 : this.meta.rating
		},
		src() {
			if (this.srcOverride) {
				return ""
			}
			let s = "small"

			if (this.scale > 0.25) {
				s = "medium"
			}
			if (this.scale > 1) {
				s = "large"
			}
			return this.thumbs[s].url
		},
		scale() {
			return parseFloat(this.display_scale)
		},
		height() {
			return this.thumbs['large'].height * this.scale
		},
		width() {
			return this.thumbs['large'].width * this.scale
		},
		...mapState('images', ['selected']),
		...mapState('interface', ['display_scale']),
	},
	methods: {
		showMenu(e) {
			e.preventDefault()

			this.menu = false
			this.menu_x = e.clientX
			this.menu_y = e.clientY
			this.$nextTick(() => {
				this.menu = true
			})
		},
		error(e) {
			if (this.src == "") {
				return
			}
			this.hasErr = true
		},
		onClick(e) {
			e.preventDefault()
			if (this.hasErr) {
				this.reload()
				return
			}
			this.$emit('click', e)
		},
		reload() {
			this.srcOverride = true
			this.hasErr = false
			setTimeout(() => {
				this.srcOverride = false
			}, 2)
		},
		...mapMutations('images', ['rate']),
	},
	components: { Rating, Tags, SizeSelect },
}
</script>

<style>
.select-check {
	position: absolute;
	top: 2px;
	right: 2px;
	background: rgba(255,255,255,0.2);
	box-shadow: 0 0 2px 0 rgba(255,255,255,0.6);
	border-radius: 50% !important;
}
</style>