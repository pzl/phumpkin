<script>
import { parseISO, format } from 'date-fns'

export default {
	props: {
		date: {},
		tz: {},
		offset: {},
	},
	computed: {
		taken() {
			if (!this.date) {
				return null
			}
			const parts = this.date.split(' ')
			parts[0] = parts[0].replace(/:/g, "-")

			let d = parts.join("T")
			
			if (this.tz) {
				d += this.tz
			} else if (this.offset) {
				d += this.offset
			}

			return format(parseISO(d), "PPpp")
		},
	},
	render(h) {
		if (this.taken === null) {
			return
		}
		return h('span', this.taken)
	}
}

</script>