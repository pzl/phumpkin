export const state = () => ({
	connected: false,
})

export const mutations = {
	setConnected(state) { state.connected = true },
	setDisconnected(state) { state.connected = false },
}