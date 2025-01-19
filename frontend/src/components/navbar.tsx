import { Component, Show } from 'solid-js'
import DarkModeToggle from './dark-mode-toggle'
import { Popover, PopoverContent, PopoverTrigger } from './ui/popover'
import { Button } from './ui/button'
import { FiLogOut, FiRss, FiUser } from 'solid-icons/fi'
import {
	DropdownMenu,
	DropdownMenuContent,
	DropdownMenuItem,
	DropdownMenuLabel,
	DropdownMenuSeparator,
	DropdownMenuTrigger,
} from './ui/dropdown-menu'
import { Avatar, AvatarFallback } from './ui/avatar'
import { useConnect } from '~/context/connect-context'
import { useNavigate } from '@solidjs/router'
import { useState } from '~/context/state-context'
import FeedConfigurator from './feed-configurator'
import { logoutAndRemoveAuthData } from '~/services/auth-service'

const Navbar: Component = () => {
	const connect = useConnect()
	const state = useState()
	const navigate = useNavigate()

	// Call this to announce that we need the user
	state.fetchUser()

	const logout = async () => {
		await connect.logout({})
		logoutAndRemoveAuthData()
		navigate('/')
	}

	return (
		<div class="flex flex-col md:flex-row items-center justify-between space-y-2">
			<div class="flex flex-col w-full justify-start">
				<h2 class="text-2xl font-bold tracking-tight">Welcome back!</h2>
				<p class="text-muted-foreground">Here&apos;s a list of your recent releases!</p>
			</div>
			<div class="flex w-full items-center justify-end space-x-4">
				<DarkModeToggle />
				<Popover>
					<PopoverTrigger as={Button<'button'>} class="cursor-pointer">
						<FiRss class="w-6 h-6" />
						Feeds
					</PopoverTrigger>
					<PopoverContent class="flex flex-col gap-4 !w-auto">
						<FeedConfigurator />
					</PopoverContent>
				</Popover>

				<DropdownMenu>
					<DropdownMenuTrigger>
						<Avatar class="size-10 cursor-pointer">
							<Show when={state.user} fallback={<AvatarFallback class="uppercase">&nbsp;</AvatarFallback>}>
								<AvatarFallback class="uppercase">{state.user?.name[0]}</AvatarFallback>
							</Show>
						</Avatar>
					</DropdownMenuTrigger>
					<DropdownMenuContent>
						<DropdownMenuLabel class="flex items-center gap-2">
							<FiUser class="w-4 h-4" />
							{state.user?.name}
						</DropdownMenuLabel>
						<DropdownMenuSeparator />
						<DropdownMenuItem class="cursor-pointer" onClick={logout}>
							<FiLogOut class="w-4 h-4" />
							Logout
						</DropdownMenuItem>
					</DropdownMenuContent>
				</DropdownMenu>
			</div>
		</div>
	)
}

export default Navbar
