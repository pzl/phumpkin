<template>
	<div>
		<v-row no-gutters>
			<v-navigation-drawer style="height: 90vh">
				<v-list-group>

				</v-list-group>
			</v-navigation-drawer>
			<photo-grid :images="[]" @more="">
				<template v-slot:before v-if="selected.length">
					<v-row class="content-group selected-labels" no-gutters>
						<p>Current Labels:</p>
						<v-chip v-for="(s,i) in selected" :key="'s'+i" close @click:close="unselect(s,i)">{{s}}</v-chip>
					</v-row>
				</template>
			</photo-grid>
			
		</v-row>
	</div>
</template>

<script>
import PhotoGrid from '~/components/photoGrid'
import { mapState, mapMutations, mapActions } from 'vuex'


function toColor(c) {
	switch (parseInt(c)) {
		case 0: return "red";
		case 1: return "yellow";
		case 2: return "green";
		case 3: return "blue";
		case 4: return "purple";
		default: return "black"
	}
}

export default {
	data() {
		return {
			labels: [0,1,2,3,4],
			current_tags: [],
		}
	},
	computed: {
		selected() {
			return this.current_tags.map(t => find(t,this.tags))
		},
		...mapState('images', ['images',  'err', 'loadMore']),
	},
	methods: {
		toColor: toColor,
		unselect(el, i) {
			this.current_tags.splice(i,1)
		},
		select(items) {
			this.current_tags = items
			// reload with current selected tags
		},
		...mapMutations('images', ['pushPath']),
		...mapActions('images', ['loadImages', 'resetImages']),
	},
	mounted() {
		//this.loadTags()
		//this.loadImages()
	},
	components: { PhotoGrid }
}
</script>


<style>

</style>