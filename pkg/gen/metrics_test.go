package gen

import "testing"

func Test_metrics(t *testing.T) {
	m := newMetrics(nil)

	mrt1 := &Measurement{FileName: "file1", Path: "path1", Created: true}
	m.Measure("pkg1", mrt1)
	mrt2 := &Measurement{FileName: "file2", Path: "path2", Created: false}
	m.Measure("pkg1", mrt2)

	t.Run("Keys", func(t *testing.T) {
		keys := m.Keys()
		if len(keys) != 1 || keys[0] != "pkg1" {
			t.Errorf("Keys() returned incorrect result: %v", keys)
		}
	})

	t.Run("Get", func(t *testing.T) {
		mrt := m.Get("pkg1")
		if len(mrt) != 2 || mrt[0] != mrt1 || mrt[1] != mrt2 {
			t.Errorf("Get() returned incorrect result: %v", mrt)
		}
	})

	t.Run("Intent", func(t *testing.T) {
		intent := m.NewIntent("pkg1", &Measurement{})
		if intent == nil || intent.key != "pkg1" || intent.parent != m {
			t.Errorf("NewIntent() returned incorrect result: %v", intent)
		}

		t.Run("Measure", func(t *testing.T) {
			mrt3 := &Measurement{FileName: "file3", Path: "path3", Created: true}
			intent.Measure(mrt3)

			mrt := m.Get("pkg1")
			if len(mrt) != 3 || mrt[0] != mrt1 || mrt[1] != mrt2 || mrt[2] != mrt3 {
				t.Errorf("Measure() did not update metrics correctly: %v", mrt)
			}
		})
	})
}
