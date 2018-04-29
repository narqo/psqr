# psqr

[![GoDoc](https://godoc.org/github.com/narqo/psqr?status.svg)](https://godoc.org/github.com/narqo/psqr)

Go implementation of [The P-Square Algorithm for Dynamic Calculation of Quantiles and Histograms Without Storing Observations][1].

> The algorithm is proposed for dynamic calculation of [..] quantiles. The estimates are produced dynamically as the observations are generated. The observations are not stored, therefore, the algorithm has a very small and fixed storage requirement regardless of the number of observations.

[1]: http://www.cs.wustl.edu/~jain/papers/ftp/psqr.pdf