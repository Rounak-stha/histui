import Hero from '@/components/hero'
import Terminal from '@/components/terminal'
import ScrollNarrative from '@/components/scroll-narrative'
import DataVizPlayground from '@/components/data-viz-playground'
import Footer from '@/components/footer'
import NavHeader from '@/components/nav-header'
import CTASection from '@/components/cta-section'
import RoadmapSection from '@/components/roadmap-section'
import UseCasesSection from '@/components/use-cases-section'
import QuickStartSection from '@/components/quick-start-section'

export default function Home() {
	return (
		<>
			<NavHeader />
			<main className='min-h-screen bg-background text-foreground overflow-hidden'>
				<Hero />
				<Terminal />
				<ScrollNarrative />
				<RoadmapSection />
				<UseCasesSection />
				<Footer />
			</main>
		</>
	)
}
