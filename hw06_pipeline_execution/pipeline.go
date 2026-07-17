package hw06pipelineexecution

type (
	In  = <-chan interface{}
	Out = In
	Bi  = chan interface{}
)

type Stage func(in In) (out Out)

func ExecutePipeline(in In, done In, stages ...Stage) Out {
	closeStage := func(in In) Out {
		if done == nil {
			return in
		}
		out := make(Bi)
		go func() {
			for {
				select {
				case <-done:
					close(out)
					for {
						_, ok := <-in
						if !ok {
							break
						}
					}
					return
				case v, ok := <-in:
					if !ok {
						close(out)
						return
					}
					out <- v
				}
			}
		}()
		return out
	}

	pipeline := func(in In) Out {
		for _, stage := range stages {
			in = closeStage(stage(in))
		}
		return in
	}(in)

	return pipeline
}
