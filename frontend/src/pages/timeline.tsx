import { Component, createEffect, createResource, createSignal, For, Show } from 'solid-js'
import { useConnect } from '~/context/connect-context'
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from '~/components/ui/card'
import { Avatar, AvatarFallback } from '~/components/ui/avatar'
import { timestampDate } from '@bufbuild/protobuf/wkt'
import { Button, buttonVariants } from '~/components/ui/button'
import { Popover, PopoverContent, PopoverTrigger } from '~/components/ui/popover'
import { Switch, SwitchControl, SwitchDescription, SwitchLabel, SwitchThumb } from '~/components/ui/switch'
import { TextField, TextFieldInput } from '~/components/ui/text-field'
import { Badge } from '~/components/ui/badge'
import CopyText from '~/components/copy-text'

const TimelinePage: Component = () => {
	const connect = useConnect()
	const [timeline] = createResource(() => connect.getRepositories({}))
	const [user, { refetch: refetchUser }] = createResource(() => connect.getMyUser({}))

	const [search, setSearch] = createSignal('')
	const [descriptionEnabled, setDescriptionEnabled] = createSignal(true)

	const filteredTimeline = () =>
		timeline()?.timeline.filter(x => x.repositoryName.toLowerCase().includes(search().toLowerCase())) ?? []

	const setUserPublic = async (isPublic: boolean) => {
		await connect.toogleUserPublicFeed({ enabled: isPublic })
		refetchUser()
	}

	return (
		<div class="flex flex-col gap-4 container pt-4">
			<div class="flex items-center justify-between space-y-2">
				<div>
					<h2 class="text-2xl font-bold tracking-tight">Welcome back!</h2>
					<p class="text-muted-foreground">Here&apos;s a list of recent releases!</p>
				</div>
				<div class="flex items-center space-x-6">
					<Popover>
						<PopoverTrigger as={Button<'button'>} class="cursor-pointer">
							RSS Feed
						</PopoverTrigger>
						<PopoverContent class="flex flex-col gap-4 !w-auto">
							<CardTitle>RSS / Atom Feed</CardTitle>
							<CardDescription class="max-w-80">
								Enable your feed as a personal <b>public</b> feed. Note that everyone who has the URL
								can has access to it (even if your profile is private).
							</CardDescription>
							<div class="grid gap-4">
								<div class="flex items-center space-x-4 rounded-md border p-4">
									<div class="flex flex-col space-y-1">
										<p class="text-sm font-medium leading-none">Enabled</p>
										<p class="text-sm text-muted-foreground">
											Get this feed as a RSS or Atom feed.
										</p>
									</div>
									<Switch checked={user()?.isPublic ?? false} onChange={e => setUserPublic(e)}>
										<SwitchControl>
											<SwitchThumb />
										</SwitchControl>
									</Switch>
								</div>

								<div class="flex w-full items-center rounded-md border p-4">
									<div class="flex w-full flex-col space-y-2 justify-items-start">
										<p class="text-sm font-medium leading-none">URLs</p>
										<p class="text-sm text-muted-foreground">
											Get this feed as a RSS or Atom feed.
										</p>
										<Badge round variant="secondary" class="text-sm mr-auto">
											RSS
										</Badge>
										<CopyText text="https://blablabla.com/rss.xml" />

										<Badge round variant="secondary" class="text-sm mr-auto">
											Atom
										</Badge>
										<CopyText text="https://blablabla.com/atom.xml" />
									</div>
								</div>
							</div>
						</PopoverContent>
					</Popover>
					<Avatar class="size-10">
						<AvatarFallback class="uppercase">{user()?.name[0]}</AvatarFallback>
					</Avatar>
				</div>
			</div>
			<div class="flex gap-4 items-center">
				<TextField>
					<TextFieldInput
						placeholder={'Search repositories'}
						value={search()}
						onInput={e => setSearch(e.currentTarget.value)}
					/>
				</TextField>

				<Switch class="items-center flex gap-2" checked={descriptionEnabled()} onChange={setDescriptionEnabled}>
					<SwitchControl>
						<SwitchThumb />
					</SwitchControl>
					<SwitchLabel>Show release description</SwitchLabel>
				</Switch>
			</div>
			<For each={filteredTimeline()}>
				{timelineItem => (
					<Card class="mx-auto max-w-120 hover:shadow-lg transition-shadow duration-200">
						<a href={timelineItem.url} target="_blank" rel="noopener noreferrer">
							<CardHeader class="!p-0">
								<img class="rounded-t-lg" src={timelineItem.imageUrl} alt={timelineItem.name} />
							</CardHeader>
							<CardContent class="flex flex-col !pb-0 pt-4 prose">
								<span class="font-normal">{timelineItem.repositoryName}</span>
								<h2 class="!mt-0 !mb-4 font-normal">{timelineItem.name}</h2>
								<Show when={descriptionEnabled()}>
									<div class="prose-sm" innerHTML={timelineItem.description}></div>
								</Show>
							</CardContent>
							<CardFooter class="text-muted-foreground !pt-2 text-sm">
								{timestampDate(timelineItem.releasedAt!).toLocaleString()}
							</CardFooter>
						</a>
					</Card>
				)}
			</For>
		</div>
	)
}

export default TimelinePage
