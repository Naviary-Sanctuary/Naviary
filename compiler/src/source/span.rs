#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub struct Span {
    start: usize,
    end: usize,
}

impl Span {
    pub fn start(self) -> usize {
        self.start
    }
    pub fn end(self) -> usize {
        self.end
    }
    pub fn new(start: usize, end: usize) -> Self {
        assert!(start <= end);

        Self { start, end }
    }

    pub fn length(self) -> usize {
        self.end - self.start
    }

    pub fn is_empty(self) -> bool {
        self.start == self.end
    }

    pub fn join(self, other: Span) -> Span {
        Self::new(self.start.min(other.start), self.end.max(other.end))
    }
}

#[cfg(test)]
mod tests {
    use super::Span;

    mod new_tests {
        use super::Span;

        #[test]
        fn creates_span_with_args() {
            let span = Span::new(3, 6);

            assert_eq!(span.start(), 3);
            assert_eq!(span.end(), 6);
        }

        #[test]
        #[should_panic]
        fn rejects_span_when_end_is_less_than_start() {
            Span::new(3, 2);
        }
    }

    mod length_tests {
        use super::Span;

        #[test]
        fn calculate_span_length() {
            let span = Span::new(1, 5);

            assert_eq!(span.length(), 4);
        }
    }

    mod is_empty_tests {
        use super::Span;

        #[test]
        fn detects_empty_span() {
            let span = Span::new(1, 1);
            assert!(span.is_empty())
        }

        #[test]
        fn detects_non_empty_span() {
            let span = Span::new(1, 2);
            assert!(!span.is_empty())
        }
    }

    mod join_tests {
        use super::Span;

        #[test]
        fn joins_two_spans() {
            let left = Span::new(0, 2);
            let right = Span::new(1, 6);

            let joined = left.join(right);

            assert_eq!(joined.start(), 0);
            assert_eq!(joined.end(), 6);
        }

        #[test]
        fn joins_spans_in_reverse_order() {
            let left = Span::new(4, 5);
            let right = Span::new(0, 1);

            let joined = left.join(right);

            assert_eq!(joined.start(), 0);
            assert_eq!(joined.end(), 5)
        }
    }
}
