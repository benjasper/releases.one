import { Component, createMemo, createResource, Show } from 'solid-js'
import DarkModeToggle from './dark-mode-toggle'
import { Popover, PopoverContent, PopoverTrigger } from './ui/popover'
import { Button } from './ui/button'
import { FiLogOut, FiRss, FiUser } from 'solid-icons/fi'
import { CardDescription, CardTitle } from './ui/card'
import CopyText from './copy-text'
import { Switch, SwitchControl, SwitchThumb } from './ui/switch'
import {
	DropdownMenu,
	DropdownMenuContent,
	DropdownMenuItem,
	DropdownMenuLabel,
	DropdownMenuSeparator,
	DropdownMenuTrigger,
} from './ui/dropdown-menu'
import { Avatar, AvatarFallback } from './ui/avatar'
import { Link } from '@solidjs/meta'
import { useConnect } from '~/context/connect-context'
import { useNavigate } from '@solidjs/router'

const Navbar: Component = () => {
	const connect = useConnect()
	const navigate = useNavigate()

	const [user, { refetch: refetchUser }] = createResource(() => connect.getMyUser({}))

	const rssFeed = createMemo(() => `${import.meta.env.VITE_API_BASE_URL}/rss/${user()?.publicId ?? ''}`)
	const atomFeed = createMemo(() => `${import.meta.env.VITE_API_BASE_URL}/atom/${user()?.publicId ?? ''}`)

	const setUserPublic = async (isPublic: boolean) => {
		await connect.toogleUserPublicFeed({ enabled: isPublic })
		refetchUser()
	}

	const logout = async () => {
		await connect.logout({})
		navigate('/login')
	}

	return (
		<div class="flex items-center justify-between space-y-2">
			<Show when={user()?.isPublic}>
				<Link rel="alternate" type="application/rss+xml" href={rssFeed()} />
				<Link rel="alternate" type="application/atom+xml" href={atomFeed()} />
			</Show>
			<div>
				<h2 class="text-2xl font-bold tracking-tight">Welcome back!</h2>
				<p class="text-muted-foreground">Here&apos;s a list of your recent releases!</p>
			</div>
			<div class="flex items-center space-x-4">
				<DarkModeToggle />
				<Popover>
					<PopoverTrigger as={Button<'button'>} class="cursor-pointer">
						<FiRss class="w-6 h-6" />
						Feeds
					</PopoverTrigger>
					<PopoverContent class="flex flex-col gap-4 !w-auto">
						<CardTitle>Feeds</CardTitle>
						<CardDescription class="max-w-80">
							Enable your feed as a personal <b>public</b> feed. Note that everyone who has the URL can
							has access to it (even if your profile is private).
						</CardDescription>
						<div class="grid gap-4">
							<div class="flex items-center space-x-4 rounded-md border p-4">
								<div class="flex flex-col space-y-1">
									<p class="text-sm font-medium leading-none">Enabled</p>
									<p class="text-sm text-muted-foreground">Get this feed as a RSS or Atom feed.</p>
								</div>
								<Switch checked={user()?.isPublic ?? false} onChange={e => setUserPublic(e)}>
									<SwitchControl>
										<SwitchThumb />
									</SwitchControl>
								</Switch>
							</div>

							<Show when={user()?.isPublic}>
								<div class="flex w-full items-center rounded-md border p-4">
									<div class="flex w-full flex-col space-y-2 justify-items-start max-w-72">
										<p class="text-sm font-medium leading-none">URLs</p>
										<p class="text-sm text-muted-foreground">
											Get this feed as a RSS or Atom feed.
										</p>
										<span class="text-sm">RSS</span>
										<CopyText text={rssFeed()} />
										<span class="text-sm">Atom</span>
										<CopyText text={atomFeed()} />
									</div>
								</div>
							</Show>
						</div>
					</PopoverContent>
				</Popover>

				<DropdownMenu>
					<DropdownMenuTrigger>
						<Avatar class="size-10 cursor-pointer">
							<Show when={user()} fallback={<AvatarFallback class="uppercase">&nbsp;</AvatarFallback>}>
								<AvatarFallback class="uppercase">{user()?.name[0]}</AvatarFallback>
							</Show>
						</Avatar>
					</DropdownMenuTrigger>
					<DropdownMenuContent>
						<DropdownMenuLabel class="flex items-center gap-2">
							<FiUser class="w-4 h-4" />
							{user()?.name}
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
