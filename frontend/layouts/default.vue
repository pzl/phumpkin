<template>
	<v-app v-scroll="onScroll">
		<!--<v-system-bar app>Phumpkin</v-system-bar>-->

		<!-- Side bar -->
		<v-navigation-drawer app v-model="nav_vis">
			<v-list-item>
				<v-list-item-content>
					<v-list-item-title class="title d-flex">
						<span>Discover</span>
						<v-spacer />
						<v-btn @click="nav_vis = false" icon class="d-block d-sm-none">
							<v-icon>mdi-close</v-icon>
						</v-btn>
					</v-list-item-title>
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
		</v-navigation-drawer>


		<!-- Right side bar -->
		<v-navigation-drawer app right :clipped="!navCollapsed" :value="infoCollapsed">
			<v-list-item>
				<v-list-item-icon>
					<v-btn small icon @click="infobar_vis = !infobar_vis">
						<v-icon>mdi-information</v-icon>
					</v-btn>
				</v-list-item-icon>
				<v-list-item-content>
					<v-list-item-title class="title">Info<template v-if="selected.length > 1"> ({{selected.length}})</template></v-list-item-title>
				</v-list-item-content>
			</v-list-item>
			<v-divider />

			<template v-if="selected.length < 4">
				<detail-card v-for="(s,i) in selected_image" :key="'summary'+i" v-bind="s" />
			</template>
			<shot-list v-else :photos="selected_image" />
		</v-navigation-drawer>

		<!-- top bar -->
		<v-app-bar app dense :collapse-on-scroll="!anySelected" :color="anySelected ? 'primary' : ''" :dark="anySelected" :clipped-right="!navCollapsed">
			<v-app-bar-nav-icon @click.stop="nav_vis = !nav_vis">
				<v-icon>mdi-{{ nav_vis ? 'backburger' : 'menu' }}</v-icon>
			</v-app-bar-nav-icon>
			<v-toolbar-title class="d-none d-md-block" >{{ anySelected ? `${selected.length} Selected` : 'Phumpkin' }}</v-toolbar-title>
			<v-btn icon v-if="!connected" class="red--text" @click="reconnect">
				<v-icon small>mdi-lan-disconnect</v-icon>
			</v-btn>
			<!--<span>{{display_scale}}</span>-->
			<v-spacer />
			<template v-if="anySelected">
				<v-btn icon @click="clearSelection">
					<v-icon>mdi-close</v-icon>
				</v-btn>
				<v-btn icon @click="lightbox = true">
					<v-icon>mdi-eye</v-icon>
				</v-btn>
				<template v-if="selected.length === 1">
					<v-btn icon :href="selected_image[0].original.url">
						<v-icon>mdi-download</v-icon>
					</v-btn>
					<v-btn icon @click="showSizeMenu">
						<v-icon>mdi-dots-vertical</v-icon>
					</v-btn>
				</template>
				<v-btn icon @click="infobar_vis = !infobar_vis">
					<v-icon>{{ infobar_vis ? 'mdi-do-not-disturb-off' : 'mdi-information' }}</v-icon>
				</v-btn>
				<size-select :x="size_menu.x" :y="size_menu.y" v-model="size_menu.show" :thumbs="selected_image.map(x=>x.thumbs)" />
				<v-spacer />
			</template>
			<!--
			<v-btn icon v-if="navCollapsed">
				<v-icon>mdi-tools</v-icon>
			</v-btn>
			-->
			<v-menu offset-y close-delay="300" :close-on-content-click="false">
				<template v-slot:activator="{ on }">
					<v-btn class="d-none d-md-block" icon v-on="on" title="Layers">
						<v-icon>mdi-layers-outline</v-icon>
					</v-btn>
				</template>
				<v-list dense>
					<v-list-item-group multiple :value="active_layers" @change="setActiveLayers">
						<v-list-item v-for="(item,i) in layers" :key="'layer'+i" :value="item.text">
							<template v-slot:default="{active,toggle}">
								<v-list-item-icon>
									<v-icon :color="active ? item.color : undefined">{{ item.icon }}</v-icon>
								</v-list-item-icon>
								<v-list-item-content>
									<v-list-item-title>{{ item.text }}</v-list-item-title>
								</v-list-item-content>
								<v-list-item-action>
									<v-checkbox :color="active ? item.color : undefined" dense :input-value="active" @click="toggle" />
								</v-list-item-action>
							</template>
						</v-list-item>
					</v-list-item-group>
				</v-list>
			</v-menu>
			<v-menu offset-y open-on-hover close-delay="300" :close-on-content-click="false">
				<template v-slot:activator="{ on }">
					<v-btn class="d-none d-md-block" icon v-on="on" title="View Size">
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
			<!--
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
			-->
			<v-btn icon @click="darkness = !darkness" title="Dark Mode" class="ml-10">
				<v-icon>mdi-theme-light-dark</v-icon>
			</v-btn>
		</v-app-bar>

		<viewer v-if="lightbox" @close="lightbox = false" :photos="selected_image" />

		<scroll-up />

		<v-content>
			<nuxt />
		</v-content>

		<v-snackbar :value="toast.show" :color="toast.style">{{ toast.message }} <v-btn dark text @click="toast.show = false">Close</v-btn></v-snackbar>

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
import Viewer from '~/components/viewer'
import DetailCard from '~/components/info/detailCard'
import ShotList from '~/components/info/shotList'
import SizeSelect from '~/components/sizeSelect'
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
			infobar_vis: false,
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
			size_menu: {
				show: false,
				x: 0,
				y: 0,
			},
		}
	},
	computed: {
		anySelected() { return !!this.$store.state.images.selected.length },
		navCollapsed() { return this.scrolled && !this.anySelected },
		infoCollapsed() { return this.infobar_vis && this.anySelected },
		selected_image() {
			return this.selected.map(i => this.images[i])
		},
		bg() { return (this.$vuetify.theme.dark) ? '#424242' : '#fafafa' },
		...mapState('images', ['images','selected','sort','sort_asc']),
		...mapState('socket',['connected']),
		...mapState('interface', ['display_scale', 'layers', 'active_layers']),
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
		sort_change(v) {
			this.sortBy(v)
			this.resetImages()
		},
		flipSortDir() {
			this.sortDir(!this.sort_asc)
			this.resetImages()
		},
		showSizeMenu(e) {
			e.preventDefault()

			this.size_menu.show = false
			this.size_menu.x = e.clientX
			this.size_menu.y = e.clientY
			this.$nextTick(() => {
				this.size_menu.show = true
			})
		},
		keyHandler(ev) {
			if (!this.lightbox && ev.keyCode === 86 && this.anySelected) {
				this.lightbox = true
			}
		},
		...mapMutations('images', ['clearSelection', 'sortBy', 'sortDir']),
		...mapMutations('interface', ['scale', 'setActiveLayers']),
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
	components: { scrollUp, Rating, Tags, DetailCard, ShotList, Viewer, SizeSelect }
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