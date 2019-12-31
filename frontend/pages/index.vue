<template>
	<v-container fluid class="d-flex flex-column">
		<v-row>
			<path-crumbs />
		</v-row>
		<v-row>
			<directory v-for="(d,i) in dirs" :key="d+i" :name="d" @click="onDirClick(d, $event)" />
		</v-row>
		<v-row justify="space-between" align="start" style="max-width: 100%">
			<thumb
				v-for="(img, i) in images" :key="i"
				v-bind="img"
				:index="i"
				@click="onClick(i, $event)"
			/>
			<div v-if="loading" class="my-12 d-flex" style="flex-basis: 100%">
				<v-progress-circular indeterminate class="mx-auto" color="deep-orange lighten-2" />
			</div>
			<span v-else class="ender" v-intersect="{ handler: intersect, options: { threshold: [0,1] }}"></span>
			<div v-if="err" class="my-12 d-flex" style="flex-basis: 100%">
				<v-btn icon class="mx-auto error--text" x-large @click="loadImages">
					<v-icon x-large>mdi-reload-alert</v-icon>
				</v-btn>
			</div>
		</v-row>
	</v-container>
</template>

<script>
import Thumb from '~/components/thumb'
import Directory from '~/components/directory'
import PathCrumbs from '~/components/pathCrumbs'
import { mapState, mapMutations, mapActions } from 'vuex'

export default {
	data() {
		return {
		}
	},
	computed: {
		...mapState('images', ['images', 'dirs', 'selected', 'loading', 'err']),
	},
	methods: {
		onClick(img, e) {
			e.preventDefault()

			// shift = add selection range
			if (e.shiftKey && this.selected.length) {
				let lastSel = this.selected[this.selected.length - 1]

				if (lastSel < img) {
					for (let i=lastSel; i<=img; i++) {
						this.addSelection(i)
					}
				} else {
					for (let i=img; i<=lastSel; i++) {
						this.addSelection(i)
					}
				}
				return
			}

			// ctrl = add image to current selection
			if (e.ctrlKey) {
				this.toggleSelect(img)
				return
			}

			// only reset selection if this is a newly clicked tile
			// this prevents losing a large range if clicking an already-selected one
			if (!this.selected.includes(img)) {
				this.setSelection(img)

			}
		},
		onDirClick(dir, e) {
			e.preventDefault()

			this.pushPath(dir)
			this.resetImages()
			this.loadImages()
		},
		intersect(entries, observer, isIntersecting) {
			if (!isIntersecting) {
				return false
			}
			if (this.loading) {
				return false
			}
			if (this.err) {
				return false
			}
			console.log("scrolled to end. loading more")
			this.loadImages()
		},
		...mapMutations('images', ['clearSelection', 'pushPath']),
		...mapActions('images', ['toggleSelect', 'setSelection', 'addSelection', 'loadImages', 'resetImages']),
	},
	mounted() {
		this.loadImages()
	},
	components: { Thumb, Directory, PathCrumbs }
}
</script>


<style>
.thumby {
	cursor: pointer;
}

.v-card--reveal {
	align-items: center;
  opacity: .5;
  bottom: 0;
  position: absolute;
  width: 100%;

  user-select: none; /* prevent name highlighting on shift-click */
}

</style>