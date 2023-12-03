import { defineStore } from 'pinia'

interface ClockState {
    currentTime: Date;
}

export const useClockStore = defineStore({
    id: 'clock',
    state: () => ({
        currentTime: new Date()
    }) as ClockState,
    actions: {
        updateTime() {
            this.currentTime = new Date()
        }
    },
})
