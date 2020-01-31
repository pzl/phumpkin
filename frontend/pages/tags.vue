<template>
	<v-row no-gutters style="flex-wrap: nowrap;">
		<v-navigation-drawer style="height: 90vh; flex-shrink: 0">
			<v-text-field v-model="tagSearch" label="find..." solo flat hide-details clearable clear-icon="mdi-close-circle-outline" prepend-inner-icon="mdi-magnify" />
			<v-treeview  ref="tree" activatable :active="current_tags" multiple-active :search="tagSearch" dense hoverable :items="tags" @update:active="select"  />
		</v-navigation-drawer>
		<photo-grid :images="images" @more="loadImages(url)">
			<template v-slot:before v-if="selected.length">
				<v-row class="content-group selected-tags" no-gutters>
					<p>Current Tags:</p>
					<v-chip v-for="(s,i) in selected.map(n=>n.join(' > '))" :key="'s'+i" close @click:close="unselect(s,i)">{{s}}</v-chip>
				</v-row>
			</template>
		</photo-grid>
		
	</v-row>
</template>

<script>
import PhotoGrid from '~/components/photoGrid'
import { mapState, mapMutations, mapActions } from 'vuex'


function merge(target, source) {
	const dest = target.slice()

	for (const x of source) {
		const idx = dest.findIndex(d=>d.name == x.name)
		if (idx === -1) {
			dest.push(x)
		} else {
			dest[idx].children = merge(dest[idx].children, x.children)
		}
	}
	return dest
}

function find(id, tags) {
	for (const t of tags) {
		if (t.id === id) {
			return [t.name]
		}
		if ('children' in t && t.children && t.children.length) {
			const child = find(id, t.children)
			if (child.length > 0) {
				return [t.name, ...child]
			}
		}
	}
	return []
}

export default {
	data() {
		return {
			tags: [],
			current_tags: [],
			tagSearch: null,

		}
	},
	computed: {
		url() { return "/api/v1/query/tags?t=" + this.selected.map(s=>s.join('|')) },
		selected() { return this.current_tags.map(t => find(t,this.tags)) },
		...mapState('images', ['images', 'err']),
	},
	methods: {
		loadTags(){
			this.$fetch('/api/v1/complete/xmp/value?field=tags')
				.then(d => {
					let id=0
					let tags  = d.data.values.map(t => 
						t.split('|')
						 .filter(x=>x!=="")
						 .reduceRight((prev,now) => {
						 	return {
						 		id:id++,
						 		name: now,
						 		children: prev? [prev] : [],
						 	}
						 }, null)
					)

					this.tags = merge([], tags)
				})
		},
		unselect(el, i) { this.current_tags.splice(i,1) },
		select(items) { this.current_tags = items },
		...mapActions('images', ['loadImages', 'resetImages']),
	},
	mounted() {
		this.loadTags()
		this.resetImages()
	},
	watch: {
		tagSearch(val, old) {
			if (!!val !== !!old) { // expand all when searching
				this.$refs.tree.updateAll(!!val)
			}
		},
		selected() {
			this.resetImages()
		}
	},
	components: { PhotoGrid }
}
</script>


<style>

</style>