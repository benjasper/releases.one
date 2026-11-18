import { Component, Match, Show, Switch } from 'solid-js'
import * as SelectPrimitive from '@kobalte/core/select'
import { RepositoryStarType, StarList } from '~/lib/generated/api/v1/api_pb'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from './ui/select'
import { FiEye, FiList, FiLock } from 'solid-icons/fi'
import { AiFillQuestionCircle, AiFillStar } from 'solid-icons/ai'
import { Tooltip, TooltipContent, TooltipTrigger } from './ui/tooltip'

export type RepositorySource = {
	starType: RepositoryStarType | null
	listId: number | null
}

type Props = RepositorySource & {
	lists: StarList[]
	onChange: (source: RepositorySource) => void
}

type Section = {
	label: string
	options: string[]
}

const toKey = (source: RepositorySource): string => {
	if (source.listId !== null) return `list:${source.listId}`
	if (source.starType === RepositoryStarType.STAR) return 'star'
	if (source.starType === RepositoryStarType.WATCH) return 'watch'
	return 'all'
}

const fromKey = (key: string): RepositorySource => {
	if (key.startsWith('list:')) return { starType: null, listId: Number(key.slice('list:'.length)) }
	if (key === 'star') return { starType: RepositoryStarType.STAR, listId: null }
	if (key === 'watch') return { starType: RepositoryStarType.WATCH, listId: null }
	return { starType: null, listId: null }
}

const RepositorySourceSelect: Component<Props> = props => {
	const sections = (): Section[] => {
		const result: Section[] = [{ label: 'Repositories', options: ['all', 'star', 'watch'] }]
		if (props.lists.length > 0) {
			result.push({ label: 'Lists', options: props.lists.map(list => `list:${list.id}`) })
		}
		return result
	}

	const findList = (key: string) => {
		const id = fromKey(key).listId
		return props.lists.find(list => list.id === id)
	}

	return (
		<div class="flex items-center gap-4 justify-between">
			<label for="repository-source" class="flex items-center text-sm font-medium leading-none">
				Repositories
				<Tooltip>
					<TooltipTrigger>
						<AiFillQuestionCircle class="ml-1 h-4 w-4 cursor-help text-muted-foreground" />
					</TooltipTrigger>
					<TooltipContent>
						<p class="text-sm text-muted-foreground max-w-72">
							You can choose to receive notifications for both watched and starred repositories, just
							watched repositories, just starred repositories, or the repositories of one of your GitHub
							star lists.
						</p>
					</TooltipContent>
				</Tooltip>
			</label>
			<Select<string, Section>
				id="repository-source"
				value={toKey(props)}
				onChange={key => props.onChange(fromKey(key ?? 'all'))}
				options={sections()}
				optionGroupChildren="options"
				disallowEmptySelection
				sectionComponent={sectionProps => (
					<SelectPrimitive.Section class="px-2 pt-2 pb-1 text-xs font-medium text-muted-foreground">
						{sectionProps.section.rawValue.label}
					</SelectPrimitive.Section>
				)}
				itemComponent={itemProps => (
					<SelectItem item={itemProps.item} class="cursor-pointer">
						<SourceLabel sourceKey={itemProps.item.rawValue} list={findList(itemProps.item.rawValue)} />
					</SelectItem>
				)}>
				<SelectTrigger aria-label="Repositories" class="cursor-pointer">
					<SelectValue<string>>
						{state => {
							const key = () => state.selectedOption() ?? toKey(props)
							return <SourceLabel sourceKey={key()} list={findList(key())} />
						}}
					</SelectValue>
				</SelectTrigger>
				<SelectContent />
			</Select>
		</div>
	)
}

const SourceLabel: Component<{ sourceKey: string; list?: StarList }> = props => {
	return (
		<div class="flex gap-1 items-center">
			<Switch>
				<Match when={props.sourceKey === 'all'}>Both</Match>
				<Match when={props.sourceKey === 'star'}>
					<AiFillStar class="w-4" /> Starred
				</Match>
				<Match when={props.sourceKey === 'watch'}>
					<FiEye class="w-4" /> Watched
				</Match>
				<Match when={props.sourceKey.startsWith('list:')}>
					<FiList class="w-4" /> {props.list?.name ?? 'List'}
					<Show when={props.list?.isPrivate}>
						<FiLock class="w-3 text-muted-foreground" />
					</Show>
				</Match>
			</Switch>
		</div>
	)
}

export default RepositorySourceSelect
