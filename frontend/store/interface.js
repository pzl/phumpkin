export const state = () => ({
	display_scale: 0.5,
	layers: [
		{ text: 'Focus Plane', icon: 'mdi-focus-field', color: 'yellow lighten-2' },
		{ text: 'AF Intent', icon: 'mdi-focus-auto', color: 'blue accent-4' },
		{ text: 'Focus Point', icon: 'mdi-image-filter-center-focus', color: 'red lighten-1' },
		{ text: 'Faces', icon: 'mdi-face-recognition', color: 'light-green' },
	],
	active_layers: ['Faces'],
})

export const mutations = {
	scale (state, val) { state.display_scale = val },
	setActiveLayers (state, val) { state.active_layers = val },

}

export const actions = {

}
