<template>
	<v-hover v-slot:default="{ hover }" open-delay="50" close-delay="150">
		<v-card :class="{ 'selected': isSelected }" :raised="isSelected" >
			<v-img
				class="thumby"
				@click.stop="$emit('click', $event)"
				:src="src"
				:lazy-src="thumbs['x-small'].url"
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
						<v-progress-circular indeterminate color="grey lighten-5"></v-progress-circular>
					</v-row>
				</template>
			</v-img>

			<v-icon v-if="isSelected" class="select-check" color="success" large>mdi-checkbox-marked-circle-outline</v-icon>

			<v-menu v-model="menu" :position-x="menu_x" :position-y="menu_y" absolute offset-y >
				<v-list>
					<v-list-item v-for="(s,i) in sizes" :key="i" :href="s.url" target="_blank" two-line dense>
						<v-list-item-content>
							<v-list-item-title>{{ s.name }}</v-list-item-title>
							<v-list-item-subtitle>{{ s.width }}x{{ s.height }}</v-list-item-subtitle>
						</v-list-item-content>
					</v-list-item>
				</v-list>
			</v-menu>

			<v-expand-transition v-if="view_size > 1 || view_size === false">
				<v-container v-if="hover" class="transition-fast-in-fast-out black darken-2 v-card--reveal white--text hidden-sm-and-down" fluid>
					<v-row dense class="d-flex justify-space-between align-center">
						<div>{{ name }}</div>
						<rating :value="display_rating" @input="rate({ image: index, rating: $event })" />
					</v-row>
					<div class="d-flex align-center">
						<v-tooltip bottom v-if="loc">
							<template v-slot:activator="{ on }">
								<v-icon dark x-small v-if="loc" v-on="on">mdi-map-marker</v-icon>
							</template>
							{{ loc.lat }}, {{ loc.lon }}
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
import { mapState, mapMutations, mapActions } from 'vuex'

export default {
	props: {
		index: {}, // not part of image, but it's place in the grid
		name: {},
		dir: {},
		size: {}, // in bytes
		rating: {},
		tags: {}, // array of strings
		xmp: {}, //
		loc: {}, // null or {lat:'', lon:''}
		thumbs: {}, // full: { url: "...", width: n, height: n}
		original: {}, //{ url: "...", width: n, height: n}
	},
	data() {
		return {
			hover_reject: false,
			menu: false,
			menu_x: 0,
			menu_y: 0,
		}
	},
	computed: {
		isSelected(){
			return this.selected.includes(this.index)
		},
		display_rating() {
			return this.hover_reject ? 0 : this.xmp.rating
		},
		src() {
			let s = "medium"
			switch (this.view_size) {
				case 1:
				case 2:
				case 3:
					s = "small"
					break
				case false:
					s = "medium"
					break
				case 4:
				case 6:
				case 12:
					s = "large"
					break
			}
			return this.thumbs[s].url
		},
		sizes() {
			return Object.keys(this.thumbs).map(s => {
				return {
					name: s,
					...this.thumbs[s]
				}
			}).sort((a,b) => a.width - b.width)
		},
		...mapState('images', ['selected']),
		...mapState('interface', ['view_size']),
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
		...mapMutations('images', ['rate']),
	},
	components: { Rating, Tags },
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