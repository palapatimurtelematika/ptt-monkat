import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'

export default function App() {
  return (
    <main className="flex min-h-svh items-center justify-center p-8">
      <Card className="w-full max-w-md">
        <CardHeader>
          <CardTitle>ptt-monkat</CardTitle>
        </CardHeader>
        <CardContent className="text-muted-foreground text-sm">
          Scaffolding up. Device list lands in stage 4.
        </CardContent>
      </Card>
    </main>
  )
}
