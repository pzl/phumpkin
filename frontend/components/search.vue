<template>
	<v-btn icon @click="focus" v-if="!visible">
		<v-icon>mdi-magnify</v-icon>
	</v-btn>
	<v-autocomplete
		v-else ref="textbox"
		rounded single-line dense outlined hide-details
		clearable
		auto-select-first
		hide-no-data no-filter
		return-object
		loader-height="4"
		:loading="loading"
		:search-input.sync="text"
		:items="items"
		v-model="model"
		@blur="onBlur"
		@click:clear="text = null"
	>
		<template v-slot:label>
			Find images <v-icon style="vertical-align: middle;">mdi-magnify</v-icon>
		</template>
		<template v-slot:item="{ item }">
			<v-list-item-avatar>
				<v-img :src="item.photo.thumbs['small'].url" :lazy-src="item.photo.thumbs['x-small'].url + '?purpose=lazysrc'" />
			</v-list-item-avatar>
			<v-list-item-content>
				<v-list-item-title><span v-for="(c,i) in item.str" :key="item.str+i" :class="{ 'search-bold': item.matches.indexOf(i) !== -1}">{{c}}</span></v-list-item-title>
				<v-list-item-subtitle><date :date="'DateTimeOriginal' in item.photo.exif ? item.photo.exif.DateTimeOriginal : item.photo.exif.FileModifyDate" /></v-list-item-subtitle>
			</v-list-item-content>
			<v-list-item-action>
				<camera-icon :make="item.photo.exif.Make" :model="item.photo.exif.Model" />
			</v-list-item-action>
		</template>
		<template v-slot:selection="{ attr, on, item, selected }">
			<v-chip v-bind="attr" :input-value="selected" color="blue-grey" class="white--text" v-on="on">{{ item.str }}</v-chip>
		</template>
		<template v-slot:append-item>
			{{ total }} results
		</template>
	</v-autocomplete>
</template>

<script>
import Date from '~/components/info/date'
import CameraIcon from '~/components/info/cameraIcon'
import { mapState, mapMutations, mapActions } from 'vuex'

// #@todo: better subtitle info. Rating, color labels, etc

export default {
	data() {
		return {
			visible: false,
			text: null,
			total: 0,
			items: [],
			model: null,
			loading: false,
		}
	},
	computed: {
		focused() {
			return !!(this.$refs.textbox && this.$refs.textbox.isFocused)
		}
	},
	methods: {
		focus(){
			this.visible = true
			this.$nextTick(() => {
				this.$refs.textbox.focus()
			})
		},
		onBlur() {
			if (!this.text && !this.model) {
				this.visible = false
			}
		},
		performSearch(val) {
			this.loading = "success"

			this.$fetch("/api/v1/complete/name?q="+val)
				.then(d => {
					this.items = d.data.results
					this.total = d.data.total
					this.loading = false
				})
		},
		keyHandler(ev) {
			if (this.focused) { // don't trigger if typing in search box
				return
			}
			if (ev.keyCode === 191 && !ev.ctrlKey && !ev.shiftKey && !ev.altKey && !ev.metaKey) { // '/' == search
				ev.preventDefault()
				this.focus()
			}
		},
		...mapMutations('images', ['setImages', 'setLoadMore']),
		...mapActions('images', ['resetImages']),
	},
	watch: {
		text(val) {
			if (val && val !== this.model) {
				this.performSearch(val)
			}
		},
		model(val) {
			if (val) {
				this.setImages([val.photo])
				this.setLoadMore(false)
			} else {
				this.setLoadMore(true)
				this.resetImages()
			}
		}
	},
	mounted() {
		window.addEventListener('keydown', this.keyHandler)
	},
	destroyed() {
		window.removeEventListener('keydown', this.keyHandler)
	},
	components: { Date, CameraIcon }
}
</script>

<style>
.search-hider.collapsed {
	width: 0px;
}

.search-bold {
	font-weight: bold;
	color: #e23f2e;
}
</style>