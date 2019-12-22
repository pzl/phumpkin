<template>
	<v-app v-scroll="onScroll">
		<!--<v-system-bar app>Phumpkin</v-system-bar>-->

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


		<!--
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

		<v-app-bar app dense :collapse-on-scroll="!anySelected" :color="anySelected ? 'primary' : ''" :dark="anySelected" :clipped-right="!navCollapsed">
			<v-app-bar-nav-icon @click.stop="nav_vis = !nav_vis" />
			<v-toolbar-title>{{ anySelected ? `${selected.length} Selected` : 'Phumpkin' }}</v-toolbar-title>
			<v-btn icon v-if="!connected" class="red--text" @click="reconnect">
				<v-icon small>mdi-lan-disconnect</v-icon>
			</v-btn>
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
				<v-btn icon>
					<v-icon>mdi-eye</v-icon>
				</v-btn>
				<v-btn icon>
					<v-icon>mdi-download</v-icon>
				</v-btn>
				<v-btn icon>
					<v-icon>mdi-dots-vertical</v-icon>
				</v-btn>
				<v-spacer />
			</template>
			<v-btn icon v-if="navCollapsed">
				<v-icon>mdi-tools</v-icon>
			</v-btn>
			<v-menu offset-y>
				<template v-slot:activator="{ on }">
					<v-btn icon v-on="on" title="View Size">
						<v-icon>mdi-apps</v-icon>
					</v-btn>
				</template>
				<v-list dense>
					<v-list-item v-for="(sz, i) in view_sizes" :key="i" @click="setViewAs(sz.size)">
						<v-list-item-content>
							<v-list-item-title v-text="sz.size"></v-list-item-title>
						</v-list-item-content>
						<v-list-item-icon>
							<v-icon v-text="sz.icon"></v-icon>
						</v-list-item-icon>
					</v-list-item>
				</v-list>
			</v-menu>
			<v-menu offset-y>
				<template v-slot:activator="{ on }">
					<v-btn icon v-on="on" title="Sort">
						<v-icon>mdi-sort</v-icon>
					</v-btn>
				</template>
				<v-list dense>
					<v-list-item-group v-model="sort_by">
						<v-list-item v-for="(sort, i) in sortables" :key="i" @click="">
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
			<v-btn icon @click="sort_asc = !sort_asc" small title="Sort Direction">
				<v-icon>mdi-sort-{{ sort_asc ? 'a' : 'de' }}scending</v-icon>
			</v-btn>
			<v-btn icon title="Filter">
				<v-icon>mdi-filter</v-icon>
			</v-btn>

			<div class="mb-n7 search-hider" :class="{ collapsed: !show_search }" >
				<v-text-field rounded single-line clearable dense solo filled prepend-icon="mdi-magnify" @click:prepend="show_search = !show_search">
					<template v-slot:label>
						Find images <v-icon style="vertical-align: middle;">mdi-magnify</v-icon>
					</template>
				</v-text-field>
			</div>
			<v-btn icon title="Upload">
				<v-icon>mdi-upload</v-icon>
			</v-btn>
			<div>
				<v-switch label="Dark Mode" v-model="darkness" hide-details />
			</div>
		</v-app-bar>

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
			sort_asc: true,
			show_search: false,
			scrolled: false,
			toast: {
				show: false,
				message: '',
				style: '',
			}
		}
	},
	computed: {
		anySelected() { return !!this.$store.state.images.selected.length },
		navCollapsed() { return this.scrolled && !this.anySelected },
		selected_image() {
			if (this.selected.length === 1) {
				return this.images[this.selected[0]]
			}
			return null
		},
		...mapState('images', ['images','selected']),
		...mapState('socket',['connected']),
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
		...mapMutations('images', ['clearSelection']),
		...mapActions('interface', ['setViewAs']),
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