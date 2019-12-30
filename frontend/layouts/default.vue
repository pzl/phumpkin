<template>
	<v-app v-scroll="onScroll">
		<!--<v-system-bar app>Phumpkin</v-system-bar>-->

		<!-- Side bar -->
		<v-navigation-drawer app v-model="nav_vis">
			<v-list-item>
				<v-list-item-content>
					<v-list-item-title class="title">Discover</v-list-item-title>
					<v-list-item-subtitle>Find and Filter photos</v-list-item-subtitle>
				</v-list-item-content>
			</v-list-item>
			<v-divider />

			<v-list dense nav>
				<v-list-item-group v-model="nav_selected" color="primary">
					<v-list-item v-for="(item, i) in nav_items" :key="i">
						<v-list-item-icon>
							<v-icon v-text="item.icon"></v-icon>
						</v-list-item-icon>
						<v-list-item-content>
							<v-list-item-title v-text="item.text"></v-list-item-title>
						</v-list-item-content>
					</v-list-item>
				</v-list-item-group>
			</v-list>

			<template v-slot:append>
				<summary-card v-if="selected.length === 1" v-bind="selected_image" />
			</template>
		</v-navigation-drawer>


		<!-- Right side bar
		<v-navigation-drawer app :clipped="!navCollapsed" :mini-variant.sync="toolbar_shrunk" expand-on-hover right v-model="toolbar_vis">
			<v-list-item>
				<v-list-item-icon>
					<v-icon @click.stop="toolbar_shrunk = !toolbar_shrunk">mdi-tools</v-icon>
				</v-list-item-icon>
				<v-list-item-content>
					<v-list-item-title class="title">Toolbar</v-list-item-title>
				</v-list-item-content>
			</v-list-item>
			<v-divider />

			<v-list dense nav>
				<v-list-item-group v-model="nav_selected" color="primary">
					<v-list-item v-for="(item, i) in nav_items" :key="i">
						<v-list-item-icon>
							<v-icon v-text="item.icon"></v-icon>
						</v-list-item-icon>
						<v-list-item-content>
							<v-list-item-title v-text="item.text"></v-list-item-title>
						</v-list-item-content>
					</v-list-item>
				</v-list-item-group>
			</v-list>
		</v-navigation-drawer>
	-->

		<!-- top bar -->
		<v-app-bar app dense :collapse-on-scroll="!anySelected" :color="anySelected ? 'primary' : ''" :dark="anySelected" :clipped-right="!navCollapsed">
			<v-app-bar-nav-icon @click.stop="nav_vis = !nav_vis" />
			<v-toolbar-title>{{ anySelected ? `${selected.length} Selected` : 'Phumpkin' }}</v-toolbar-title>
			<v-btn icon v-if="!connected" class="red--text" @click="reconnect">
				<v-icon small>mdi-lan-disconnect</v-icon>
			</v-btn>
			<span>{{display_scale}}</span>
			<v-spacer />
			<template v-if="anySelected">
				<v-btn icon @click="clearSelection">
					<v-icon>mdi-close</v-icon>
				</v-btn>
				<template v-if="selected.length === 1">
					<v-btn icon>
						<v-icon>mdi-information</v-icon>
					</v-btn>
				</template>
				<v-btn icon @click="view">
					<v-icon>mdi-eye</v-icon>
				</v-btn>
				<v-btn icon disabled>
					<v-icon>mdi-download</v-icon>
				</v-btn>
				<v-btn icon disabled>
					<v-icon>mdi-dots-vertical</v-icon>
				</v-btn>
				<v-spacer />
			</template>
			<v-btn icon v-if="navCollapsed">
				<v-icon>mdi-tools</v-icon>
			</v-btn>
			<v-menu offset-y open-on-hover close-delay="300" :close-on-content-click="false">
				<template v-slot:activator="{ on }">
					<v-btn icon v-on="on" title="View Size">
						<v-icon>mdi-apps</v-icon>
					</v-btn>
				</template>
				<v-slider :value="display_scale" @input="scale" hint="Size" dense max="2" min="0.03" step="0.01" vertical hide-details :background-color="bg" />
			</v-menu>
			<v-menu offset-y>
				<template v-slot:activator="{ on }">
					<v-btn icon v-on="on" title="Sort">
						<v-icon>mdi-sort</v-icon>
					</v-btn>
				</template>
				<v-list dense>
					<v-list-item-group v-model="sort_by">
						<v-list-item v-for="(sort, i) in sortables" :key="i" @click="sort_change(sort.text)">
							<v-list-item-content>
								<v-list-item-title v-text="sort.text"></v-list-item-title>
							</v-list-item-content>
							<v-list-item-icon>
								<v-icon v-text="sort.icon"></v-icon>
							</v-list-item-icon>
						</v-list-item>
					</v-list-item-group>
				</v-list>
			</v-menu>
			<v-btn icon @click="flipSortDir" small title="Sort Direction">
				<v-icon>mdi-sort-{{ sort_asc ? 'a' : 'de' }}scending</v-icon>
			</v-btn>
			<v-btn icon title="Filter" disabled>
				<v-icon>mdi-filter</v-icon>
			</v-btn>

			<div class="mb-n7 search-hider" :class="{ collapsed: !show_search }" >
				<v-text-field rounded single-line clearable dense solo filled prepend-icon="mdi-magnify" @click:prepend="show_search = !show_search" disabled>
					<template v-slot:label>
						Find images <v-icon style="vertical-align: middle;">mdi-magnify</v-icon>
					</template>
				</v-text-field>
			</div>
			<v-btn icon title="Upload" disabled>
				<v-icon>mdi-upload</v-icon>
			</v-btn>
			<div>
				<v-switch label="Dark Mode" v-model="darkness" hide-details />
			</div>
		</v-app-bar>

		<v-overlay v-if="lightbox" :value="lightbox" z-index="99">
			<v-sheet height="100%" width="100%">
				<v-btn icon @click="lightbox = false"><v-icon>mdi-close</v-icon></v-btn>
				<v-img v-if="selected.length === 1" :height="selected_image[0].thumbs['large'].height" :width="selected_image[0].thumbs['large'].width" :lazy-src="selected_image[0].thumbs['x-small'].url+'?purpose=lazysrc'" :src="selected_image[0].thumbs['large'].url+'?purpose=viewer'" />
				<v-carousel v-else v-model="lightbox_position" show-arrows-on-hover height="100%" style="width: 100%">
					<v-carousel-item v-for="(img,i) in selected_image" :key="i">
						<v-img :height="img.thumbs['large'].height" :width="img.thumbs['large'].width" :lazy-src="img.thumbs['x-small'].url+'?purpose=lazysrc'" :src="img.thumbs['large'].url+'?purpose=viewer'" />
					</v-carousel-item>
				</v-carousel>
			</v-sheet>
		</v-overlay>

		<scroll-up />

		<v-content>
			<nuxt />
		</v-content>

		<v-snackbar :value="toast.show" :color="toast.style">{{ toast.message }} <v-btn dark text @click="toast.show = false">Close</v-btn></v-snackbar>

		<v-bottom-navigation class="hidden-md-and-up" app>
		</v-bottom-navigation>

		<v-footer class="d-flex justify-space-between" app>
			<span>Phumpkin</span>
			<span class="copy">v. {{ version }} &copy; {{ new Date().getFullYear() }}</span>
		</v-footer>
	</v-app>
</template>

<script>
import scrollUp from '~/components/scrollUp'
import Rating from '~/components/rating'
import Tags from '~/components/tags'
import SummaryCard from '~/components/summaryCard'
import { mapState, mapMutations, mapActions } from 'vuex'

export default {
	data() {
		return {
			darkness: false,
			version: 'ef456d2',
			nav_vis: null,
			nav_selected: 0,
			nav_items: [
				{ text: 'Photos', icon: 'mdi-image' },
				{ text: 'Faces', icon: 'mdi-face' },
				{ text: 'Tags', icon: 'mdi-tag' },
				{ text: 'Places', icon: 'mdi-map-marker' },
			],
			toolbar_vis: null,
			toolbar_shrunk: true,
			view_sizes: [
				{size: 'auto', icon: 'mdi-collage' },
				{size: 'x-small', icon: 'mdi-drag-horizontal' },
				{size: 'small', icon: 'mdi-view-comfy' },
				{size: 'medium', icon: 'mdi-view-module' },
				{size: 'medium-pad', icon: 'mdi-apps' },
				{size: 'large', icon: 'mdi-view-grid-outline' },
				{size: 'x-large', icon: 'mdi-view-grid' },
				{size: 'single', icon: 'mdi-selection' },
			],
			sortables: [
				{ text: 'Rating', icon: 'mdi-star-half' },
				{ text: 'Date Taken', icon: 'mdi-calendar-clock' },
				{ text: 'Name', icon: 'mdi-sort-alphabetical' },
			],
			sort_by: 0,
			show_search: false,
			scrolled: false,
			toast: {
				show: false,
				message: '',
				style: '',
			},
			lightbox: false,
			lightbox_position: 0,
		}
	},
	computed: {
		anySelected() { return !!this.$store.state.images.selected.length },
		navCollapsed() { return this.scrolled && !this.anySelected },
		selected_image() {
			return this.selected.map(i => this.images[i])
		},
		bg() { return (this.$vuetify.theme.dark) ? '#424242' : '#fafafa' },
		...mapState('images', ['images','selected','sort','sort_asc']),
		...mapState('socket',['connected']),
		...mapState('interface', ['display_scale']),
	},
	methods: {
		onScroll() {
			if (typeof window === 'undefined') {
				return
			}
			const top = ( window.pageYOffset || document.documentElement.offsetTop || 0)
			this.scrolled = top > 0
		},
		reconnect() { this.$sock.reconnect() },
		view() {
			this.lightbox = true
			this.lightbox_position = 0
		},
		closeView() {
			this.lightbox = false
			this.lightbox_position = 0
		},
		sort_change(v) {
			this.sortBy(v)
			this.resetImages()
		},
		flipSortDir() {
			this.sortDir(!this.sort_asc)
			this.resetImages()
		},
		keyHandler(ev) {
			if (!this.lightbox) {
				switch (ev.keyCode) {
					case 86:
						if (this.anySelected) {
							this.view()
						}
						break
				}
				return
			}
			switch (ev.keyCode) {
				case 27: // esc
					this.closeView()
					break
				case 37: // left
					this.lightbox_position--
					break
				case 39: // right
					this.lightbox_position++
					break
			}
		},
		...mapMutations('images', ['clearSelection', 'sortBy', 'sortDir']),
		...mapMutations('interface', ['scale']),
		...mapActions('images', ['resetImages'])
	},
	watch: {
		darkness(val) { this.$vuetify.theme.dark = val },
		connected(val) {
			if (!val) {
				this.toast.show = true
				this.toast.message = "Disconnected from server"
				this.toast.style = "error"
			}
		},
	},
	mounted() {
		window.addEventListener('keydown', this.keyHandler)
	},
	destroyed() {
		window.removeEventListener('keydown', this.keyHandler)
	},
	components: { scrollUp, Rating, Tags, SummaryCard }
}
</script>


<style>

.search-hider.collapsed {
	width: 2%;
}

.search-hider.collapsed .v-input__slot {
	padding: 0;
}

</style>