package dga

import (
	"github.com/sirupsen/logrus"
	tf "github.com/tensorflow/tensorflow/tensorflow/go"
)

const MAX_LEN = 75

type DGAModel struct {
	model *tf.SavedModel
}

func New(modelpath string) *DGAModel {
	model, err := tf.LoadSavedModel(modelpath, []string{"cisco"}, nil)
	if err != nil {
		logrus.Fatal("Error loading saved model:", err.Error())
	}
	return &DGAModel{
		model: model,
	}
}

//SeqPadding : function to convert dns string to float32 array with left padding
func SeqPadding(dns string) [MAX_LEN]float32 {
	var X [MAX_LEN]float32
	namestr := []byte(dns)
	strlen := len(namestr)
	if strlen >= MAX_LEN {
		namestr = namestr[strlen-MAX_LEN : strlen]
		for idx := 0; idx < MAX_LEN; idx++ {
			X[idx] = float32(namestr[idx])
			if X[idx] >= 128 {
				X[idx] = 0
			}
		}
	} else {
		idy := 0
		for idx := MAX_LEN - strlen; idx < MAX_LEN; idx++ {
			X[idx] = float32(namestr[idy])
			if X[idx] >= 128 {
				X[idx] = 0
			}
			idy++
		}
	}
	return X
}

//Predict :main function to predict domain risk
func (d *DGAModel) Predict(dns string) bool {
	//normally dynamic domain needs more than 5 chars.
	if len(dns) < 5 {
		return true
	}
	X := SeqPadding(dns)
	tensors, _ := tf.NewTensor([][MAX_LEN]float32{X})
	r, err := d.model.Session.Run(
		map[tf.Output]*tf.Tensor{
			d.model.Graph.Operation("inputlayer_input").Output(0): tensors,
		},
		[]tf.Output{
			d.model.Graph.Operation("outputlayer/Softmax").Output(0),
		},
		nil,
	)

	if err != nil {
		logrus.Fatal(err)
	} else {
		rlist := r[0].Value().([][]float32)
		if rlist[0][0] < 0.1 {
			return true
		}
	}
	return false
}
