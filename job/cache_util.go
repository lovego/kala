package job

import "errors"

func enable(j *Job, cache JobCache, persist bool) error {
	j.lock.Lock()
	defer j.lock.Unlock()

	shouldStartWaiting := j.jobTimer != nil && j.Disabled
	j.Disabled = false

	if persist {
		j.lock.Unlock()
		if err := cache.Set(j); err != nil {
			j.lock.Lock()
			j.Disabled = true
			return err
		}
		j.lock.Lock()
	}

	if shouldStartWaiting {
		go j.StartWaiting(cache, false)
	}

	return nil
}

func disable(j *Job, cache JobCache, persist bool) error {
	if j.IsDone {
		return errors.New("done job can not disable")
	}
	j.lock.Lock()
	defer j.lock.Unlock()

	j.Disabled = true
	if persist {
		j.lock.Unlock()
		if err := cache.Set(j); err != nil {
			j.lock.Lock()
			j.Disabled = false
			return err
		}
		j.lock.Lock()
	}

	if j.jobTimer != nil {
		j.jobTimer.Stop()
	}

	return nil
}
